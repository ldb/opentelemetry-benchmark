import math
from datetime import datetime
import matplotlib as mpl
from matplotlib import pyplot
import fileinput
import numpy as np

# This script plots the rate of sending traces per second and the sending latency in ms
# X Axis: time
# Y Axis: traces per second
# Y2 Axis: receive latency in ms

# RESAMPLE_SECONDS describes what window the data should be resampled into. 
# This is necessary because the logs are not uniformly distributed, making it impossible to calculate a moving average.
# For resampling, all timestamps will be rounded UP to the nearest RESAMPLE_SECONDS seconds.
RESAMPLE_SECONDS = 1

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
    cumsum = np.cumsum(np.insert(x, 0, 0)) 
    return (cumsum[N:] - cumsum[:-N]) / float(N)

clients = {}
rate ={}
first = -1
firstError = -1
title = ""
latencyR = {}

# example: W basic-50 16:50:02.869689 0 0 0 0 0 0 0 -62135596800000 -62135596800000 -62135596800000 -62135596800000 0
for line in fileinput.input():
    try:
        kind, name, ts, id, statusCode, traceDepth, riskyAttributeDepth, extraAttributes, spanLength, coolDown, startTS, sendTS, endSendTS, receiveTS, sendReceiveDelta = line.split(" ")
    except ValueError:
        # Manager logs are shorter, leaving Python to fail unpacking
        continue

    if title == "":
        title = name

    timestamp = datetime.strptime(ts, '%H:%M:%S.%f').timestamp()

    s = resample(timestamp, RESAMPLE_SECONDS)
    if first == -1:
        first = s
    seconds = s - first
    if clients.get(seconds) is None:
        clients[seconds] = 0
    if rate.get(seconds) is None:
        rate[seconds] = 0
        latencyR[seconds] = [0]

    if statusCode == '0': # Started worker
        clients[seconds] += 1
    elif statusCode == '1': # Successfully sent a trace
        rate[seconds] += 1 / RESAMPLE_SECONDS
        latencyR[seconds].append(int(endSendTS) - int(sendTS))

    elif statusCode in ['2', '3','4'] and firstError == -1: # Any kind of error
        firstError = timestamp - first
    

rate_mean = running_mean( np.array(list(rate.values())), 20)
latency_50 = running_mean( np.array([np.percentile(l,50) for l in list(latencyR.values())]), 20)
latency_95 = running_mean( np.array([np.percentile(l,95) for l in list(latencyR.values())]), 20)
latency_99 = running_mean( np.array([np.percentile(l,99) for l in list(latencyR.values())]), 20)

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
ax.set_title("\"" + title + "\" Sending rate and send latency")

c = list(clients.values())
ax.plot(list(rate.keys())[9:-10], rate_mean, color="orange", label="traces/s")
ax.set_ylabel("traces/s", color="orange")
ax2 = ax.twinx()

ax2.set_ylabel("send latency in ms")
ax2.plot(list(rate.keys())[9:-10], latency_50, color="green", label="50th percentile")
ax2.plot(list(rate.keys())[9:-10], latency_95, color="red", label="95th percentile")
ax2.plot(list(rate.keys())[9:-10], latency_99, color="blue", label="99th percentile")
ax2.legend(loc='best', fancybox=True)


ax.axvline(x=firstError, color='red',linestyle="dotted")

#pyplot.xticks(np.arange(min(rate.keys()), max(rate.keys())+1, 180))

fig.tight_layout()
#pyplot.show()
pyplot.savefig("benchd_"+title+"_send_rate_send_latency.png")
