import math
from datetime import datetime
import matplotlib as mpl
from matplotlib import pyplot
import fileinput
import numpy as np
import os

# This script plots the number of active workers in benchd and the rate of successfully sent traces (resampled to RESAMPLE_SECONDS) per second over time.
# Resampling is done by taking the average of all values in the resample period.
# X Axis: time
# Y Axis: #clients
# Y2 Axis: traces sent per second

RESAMPLE_SECONDS = 1
MOVING_WINDOW_SECONDS = 10

if os.environ.get('BENCH_MA_WINDOW') is not None:
    print("moving average window is", os.environ.get('BENCH_MA_WINDOW'))
    MOVING_WINDOW_SECONDS = int(os.environ.get('BENCH_MA_WINDOW'))

def figureSize(width, fraction=1):
    # Set aesthetic figure dimensions to avoid scaling in latex
    fig_width_pt = width * fraction
    inches_per_pt = 1 / 72.27
    # Golden ratio to set aesthetic figure height
    golden_ratio = (5 ** 0.5 - 1) / 2
    fig_width_in = fig_width_pt * inches_per_pt
    fig_height_in = fig_width_in * golden_ratio
    return fig_width_in, fig_height_in

# Resampling
def resample(x, seconds):
    return int(math.ceil(x / float(seconds))) * seconds

# https://stackoverflow.com/questions/13728392/moving-average-or-running-mean
def running_mean(x, N):
    return np.convolve(x, np.ones(N) / float(N), 'valid')
    #cumsum = np.cumsum(np.insert(x, 0, 0)) 
    #return (cumsum[N:] - cumsum[:-N]) / float(N)

rate ={}
first = {}
firstError = -1
title = ""

cache = False

if os.environ.get('BENCH_USE_CACHE') == "true":
    cache = True
    print("using cache")
    for line in fileinput.input():
        name, seconds, value = line.split(" ")
        if rate.get(name) is None:
            rate[name] = {}
        if rate[name].get(int(seconds)) is None:
            rate[name][int(seconds)] = float(value)
else:
    # example: W basic-50 16:50:02.869689 0 0 0 0 0 0 0 -62135596800000 -62135596800000 -62135596800000 -62135596800000 0
    for line in fileinput.input():
        try:
            kind, name, ts, id, statusCode, traceDepth, riskyAttributeDepth, extraAttributes, spanLength, coolDown, startTS, sendTS, endSendTS, receiveTS, sendReceiveDelta = line.split(" ")
        except ValueError:
            # Manager logs are shorter, leaving Python to fail unpacking
            continue

        # Double parsing to get rid of the hours of the timestamp
        timestamp = datetime.strptime(ts, '%H:%M:%S.%f').timestamp()

        s = resample(timestamp, RESAMPLE_SECONDS)
        if first.get(name) is None:
            first[name] = s
        seconds = s - first[name]
        if rate.get(name) is None:
            rate[name] = {}
        if rate[name].get(seconds) is None:
            rate[name][seconds] = 0

        if statusCode in ['2', '3','4', '5'] and firstError == -1: # Any kind of error or worker exited
            continue
        elif statusCode == '1': # Successfully sent a trace
            rate[name][seconds] += 1 / RESAMPLE_SECONDS



# SIZES FOR PAPER
# pyplot.rcParams["font.family"] = "serif"
# pyplot.rcParams["mathtext.fontset"] = "dejavuserif"
# pyplot.rcParams["font.size"] = 12
# pyplot.rcParams["lines.linewidth"] = 1
# pyplot.rcParams["axes.labelsize"] = 10
# pyplot.rcParams["legend.fontsize"] = 8
# pyplot.rcParams["xtick.labelsize"] = 8
# pyplot.rcParams["ytick.labelsize"] = 8
# figsize = figureSize(450)

# SIZES FOR PRESENTATION
pyplot.rcParams["font.family"] = "serif"
pyplot.rcParams["mathtext.fontset"] = "dejavuserif"
pyplot.rcParams["font.size"] = 24
pyplot.rcParams["lines.linewidth"] = 2
pyplot.rcParams["axes.labelsize"] = 24
pyplot.rcParams["legend.fontsize"] = 24
pyplot.rcParams["xtick.labelsize"] = 24
pyplot.rcParams["ytick.labelsize"] = 24
figsize = figureSize(450, 3)

fig, ax = pyplot.subplots(sharex=True, sharey=False, figsize=figsize)

formatter = mpl.ticker.FuncFormatter(lambda s, x: datetime.utcfromtimestamp(s).strftime('%M:%S'))
ax.xaxis.set_major_formatter(formatter)
ax.set_xlabel("Runtime in minutes")
ax.set_title("Sustained throughput")

ax.set_ylabel("traces/s")

for name, rates in rate.items():
    if cache is False: # We are reading from cache, no need to write it again
        for s, v in rates.items():
            print(name, s, v)
    rate_mean = running_mean( np.array(list(rates.values())), MOVING_WINDOW_SECONDS)
    ax.plot(list(rates.keys())[:len(rate_mean)], list(rate_mean), label=name)

ax.legend(loc='best', fancybox=True)

#pyplot.xticks(np.arange(min(rate.keys()), max(rate.keys()), 120))

fig.tight_layout()
#pyplot.show()
pyplot.savefig("benchd_"+ str(MOVING_WINDOW_SECONDS)+"_sustained_throughput.png")
