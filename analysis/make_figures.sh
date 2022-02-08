#!/bin/bash

echo "\033[0;36m THIS WILL NOW TAKE A WHILE...\033[0m"

# basic-1-verify
cat results/basic-1-verify/log* | python3 analysis/benchd_number_clients_send_rate.py &
cat results/basic-1-verify/log* | python3 analysis/benchd_send_rate_receive_latency.py &
cat results/basic-1-verify/log* | python3 analysis/benchd_send_rate_send_latency.py &

# basic-100
cat results/basic-100/log* | python3 analysis/benchd_number_clients_send_rate.py &
cat results/basic-100/log* | python3 analysis/benchd_send_rate_receive_latency.py &
cat results/basic-100/log* | python3 analysis/benchd_send_rate_send_latency.py &

# basic-50
cat results/basic-50/log* | python3 analysis/benchd_number_clients_send_rate.py &
cat results/basic-50/log* | python3 analysis/benchd_send_rate_receive_latency.py &
cat results/basic-50/log* | python3 analysis/benchd_send_rate_send_latency.py &
cat results/basic-50-sustain/log* | python3 analysis/benchd_send_rate_receive_latency.py &

# realistic-50
cat results/realistic-50/log* | python3 analysis/benchd_number_clients_send_rate.py &
cat results/realistic-50/log* | python3 analysis/benchd_send_rate_receive_latency.py &
cat results/realistic-50/log* | python3 analysis/benchd_send_rate_send_latency.py &
cat results/realistic-50-sustain/log* | python3 analysis/benchd_send_rate_receive_latency.py &

# mutate-50
cat results/mutate-50/log* | python3 analysis/benchd_number_clients_send_rate.py &
cat results/mutate-50/log* | python3 analysis/benchd_send_rate_receive_latency.py &
cat results/mutate-50/log* | python3 analysis/benchd_send_rate_send_latency.py & 
cat results/mutate-50-sustain/log* | python3 analysis/benchd_send_rate_receive_latency.py &

# mutate-100-sustain
cat results/mutate-100-sustain/log* | python3 analysis/benchd_send_rate_receive_latency.py &

# sample-100
cat results/sample-100/log* | python3 analysis/benchd_number_clients_send_rate.py &
cat results/sample-100/log* | python3 analysis/benchd_send_rate_receive_latency.py &
cat results/sample-100/log* | python3 analysis/benchd_send_rate_send_latency.py &
cat results/sample-100-sustain/log* | python3 analysis/benchd_send_rate_receive_latency.py &


# Sustain

FILE=results/analysis_cache
if [ -f "$FILE" ]; then
    cat $FILE | BENCH_USE_CACHE=true BENCH_MA_WINDOW=10 python3 analysis/benchd_sustained_throughput.py &
    cat $FILE | BENCH_USE_CACHE=true BENCH_MA_WINDOW=30 python3 analysis/benchd_sustained_throughput.py &
    cat $FILE | BENCH_USE_CACHE=true BENCH_MA_WINDOW=60 python3 analysis/benchd_sustained_throughput.py &
    cat $FILE | BENCH_USE_CACHE=true BENCH_MA_WINDOW=120 python3 analysis/benchd_sustained_throughput.py &
else 
    cat results/basic-50-sustain/log* results/realistic-50-sustain/log* results/mutate-50-sustain/log* | python3 analysis/benchd_sustained_throughput.py > $FILE && {
    cat $FILE | BENCH_USE_CACHE=true BENCH_MA_WINDOW=30 python3 analysis/benchd_sustained_throughput.py &
    cat $FILE | BENCH_USE_CACHE=true BENCH_MA_WINDOW=60 python3 analysis/benchd_sustained_throughput.py &
    cat $FILE | BENCH_USE_CACHE=true BENCH_MA_WINDOW=120 python3 analysis/benchd_sustained_throughput.py &
}
fi

wait