from cProfile import label
import math
from datetime import datetime, time, timedelta
from turtle import color
import matplotlib as mpl
from matplotlib import pyplot
import fileinput
from matplotlib import style
import matplotlib.dates as md
import matplotlib.ticker as ticker
from ipaddress import *
import numpy as np
import matplotlib.dates as matdates

# This script plots the number of worker clients in benchd over time
# X Axis: time
# Y Axis: #clients
# Y2 Axis: traces sent per second

# RESAMPLE_SECONDS describes how much data should be resampled to reduce the noise.
# For resampling, all timestamps will be rounded UP to the nearest RESAMPLE_SECONDS seconds.
RESAMPLE_SECONDS = 30

def figureSize(width, fraction=1):
    # Set aesthetic figure dimensions to avoid scaling in latex
    fig_width_pt = width * fraction
    inches_per_pt = 1 / 72.27
    # Golden ratio to set aesthetic figure height
    golden_ratio = (5 ** 0.5 - 1) / 2
    fig_width_in = fig_width_pt * inches_per_pt
    fig_height_in = fig_width_in * golden_ratio
    return fig_width_in, fig_height_in

# Simple Moving Average
def movingAverage(N,list):
    padded = np.pad(list, (N // 2, N - 1 - N // 2), mode='edge')
    smooth = np.convolve(padded, np.ones((N,)) / N, mode='valid')
    return smooth

# Resampling
def resample(x, seconds):
    return int(math.ceil(x / float(seconds))) * seconds


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

    timestamp = datetime.strptime(ts, '%H:%M:%S.%f')
    s = resample(timestamp.timestamp(), RESAMPLE_SECONDS)
    if first == -1:
        first = s
    seconds = s - first
    if clients.get(seconds) is None:
        clients[seconds] = 0
    if rate.get(seconds) is None:
        rate[seconds] = 0

    if statusCode == '0': # Started worker
        clients[seconds] += 1
    elif statusCode == '1': # Suuccessfully sent a trace
        rate[seconds] += 1
    elif statusCode in ['2', '3','4'] and firstError == -1: # Any kind of error
        firstError = timestamp.timestamp() - first


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
ax.set_title("\"" + title + "\" Clients and Sending rate")

ax.set_ylabel("Active workers", color="blue")
c = list(clients.values())
ax.plot(rate.keys(), [sum(c[:y]) for y in range(1, len(c) + 1)], color="blue", label="Active workers")
ax2 = ax.twinx()
ax2.plot(rate.keys(), [r / RESAMPLE_SECONDS for r in list(rate.values())], color="orange", label="traces/s")
ax2.set_ylabel("traces/s", color="orange")

ax.axvline(x=firstError, color='red',linestyle="dotted")

pyplot.xticks(np.arange(min(rate.keys()), max(rate.keys())+1, 120))

fig.tight_layout()
pyplot.show()
