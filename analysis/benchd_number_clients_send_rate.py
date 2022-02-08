import math
from datetime import datetime
import matplotlib as mpl
from matplotlib import pyplot
import fileinput
import numpy as np

# This script plots the number of active workers in benchd and the rate of successfully sent traces (resampled to RESAMPLE_SECONDS) per second over time.
# Resampling is done by taking the average of all values in the resample period.
# X Axis: time
# Y Axis: #clients
# Y2 Axis: traces sent per second

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
        #rateMA[seconds] = 0                

    if statusCode == '0': # Started worker
        clients[seconds] += 1
    elif statusCode == '1': # Successfully sent a trace
        rate[seconds] += 1 / RESAMPLE_SECONDS # Evenly distribute the values over the RESAMPLE_SECONDS period.
        #Here, avg is (1 / #buckets) where #buckets is the number of buckets used for the moving average calculation.
        #rateMA[seconds] += avg / RESAMPLE_SECONDS
    elif statusCode in ['2', '3','4'] and firstError == -1: # Any kind of error
        firstError = timestamp - first

rate_mean = running_mean( np.array(list(rate.values())), 10)


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
ax.set_title("\"" + title + "\" Workers and Sending rate")

ax.set_ylabel("Active workers", color="blue")
c = list(clients.values())
ax.plot(rate.keys(), [sum(c[:y]) for y in range(1, len(c) + 1)], color="blue", label="Active workers", marker="*")
ax2 = ax.twinx()
ax2.plot(list(rate.keys())[4:-5] ,list(rate_mean), color="orange", label="traces/s", marker="") 
#ax2.plot(rateMA.keys(),  list(rateMA.values()), color="green", label="traces/s", marker="x") # Uncomment to see the values with moving average
ax2.set_ylabel("traces/s", color="orange")

ax.axvline(x=firstError, color='red',linestyle="dotted")

#pyplot.xticks(np.arange(min(rate.keys()), max(rate.keys())+1, 120))

fig.tight_layout()
#pyplot.show()
pyplot.savefig("benchd_"+title+"_number_clients_send_rate.png")
