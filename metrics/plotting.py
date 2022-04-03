import seaborn as sns
import pandas
import matplotlib.pyplot as plt
import numpy as np
import json
import argparse
import sys

plt.rc('text', usetex=True)
plt.rcParams['font.size'] = 24
plt.rcParams['axes.linewidth'] = 2
plt.rcParams.update({'figure.autolayout': True})
plt.rcParams.update({'font.size': 24})

def reject_outliers(data, m=2):
    return data[abs(data - np.mean(data)) < m * np.std(data)]

def aggregation_plots(path, data, title):
    fig = plt.figure(figsize=(20,7))
    ax = sns.lineplot(data=data, y="processing-time", x="frame-number", palette='pastel', linewidth=1)
    ax.axhline(data["processing-time"].mean(), linestyle='--', color='black')
    ax.set(xlabel='', ylabel='Processing Time', title=title)
    ax.yaxis.grid()
    fig.savefig(path + '/aggregation-processing-time.pdf')
    fig.savefig(path + '/aggregation-processing-time.png')
    plt.close()
    
    fig = plt.figure(figsize=(20,7))
    ax = sns.lineplot(data=data, y="skipped-frames", x="frame-number", palette='pastel', linewidth=1)
    ax.set(xlabel='', ylabel='Skipped Frames', title=title)
    ax.yaxis.grid()
    fig.savefig(path + '/aggregation-skipped-frames.pdf')
    fig.savefig(path + '/aggregation-skipped-frames.png')
    plt.close()
    
def tracking_plots(path, data, title):
    fig = plt.figure(figsize=(20,7))
    ax = sns.lineplot(data=data, y="processing-time", x="frame-number", palette='pastel', linewidth=1)
    ax.axhline(data["processing-time"].mean(), linestyle='--', color='black')
    ax.set(xlabel='', ylabel='Processing Time', title=title)
    ax.yaxis.grid()
    fig.savefig(path + '/tracking-processing-time.pdf')
    fig.savefig(path + '/tracking-processing-time.png')
    plt.close()
    
    fig = plt.figure(figsize=(20,7))
    ax = sns.lineplot(data=data, y="tracked-objects", x="frame-number", palette='pastel', linewidth=1)
    ax.axhline(data["tracked-objects"].mean(), linestyle='--', color='black')
    ax.set(xlabel='', ylabel='Tracked Objects', title=title)
    ax.yaxis.grid()
    fig.savefig(path + '/tracking-tracked-objects.pdf')
    fig.savefig(path + '/tracking-tracked-objects.png')
    plt.close()

def detection_plots(path, data, title):
    fig = plt.figure(figsize=(20,7))
    ax = sns.lineplot(data=data, y="processing-time", x="frame-number", palette='pastel', linewidth=1)
    ax.axhline(data["processing-time"].mean(), linestyle='--', color='black')
    ax.set(xlabel='', ylabel='Processing Time', title=title)
    ax.yaxis.grid()
    fig.savefig(path + '/detection-processing-time.pdf')
    fig.savefig(path + '/detection-processing-time.png')
    plt.close()
    
    fig = plt.figure(figsize=(20,7))
    ax = sns.lineplot(data=data, y="detected-objects", x="frame-number", palette='pastel', linewidth=1)
    ax.axhline(data["detected-objects"].mean(), linestyle='--', color='black')
    ax.set(xlabel='', ylabel='Tracked Objects', title=title)
    ax.yaxis.grid()
    fig.savefig(path + '/detection-detected-objects.pdf')
    fig.savefig(path + '/detection-detected-objects.png')
    plt.close()

def pipeline_result_plots(path, data):
    fig = plt.figure(figsize=(20,7))
    ax = sns.lineplot(data=data, y="misses", x="frame-number", palette='pastel', linewidth=1)
    ax.axhline(data["misses"].mean(), linestyle='--', color='black')
    ax.set(xlabel='', ylabel='Misses')
    ax.yaxis.grid()
    fig.savefig(path + '/misses.pdf')
    fig.savefig(path + '/misses.png')
    plt.close()
    
    fig = plt.figure(figsize=(20,7))
    ax = sns.lineplot(data=data, y="falsePositives", x="frame-number", palette='pastel', linewidth=1)
    ax.axhline(data["falsePositives"].mean(), linestyle='--', color='black')
    ax.set(xlabel='', ylabel='False Positives')
    ax.yaxis.grid()
    fig.savefig(path + '/falsePositives.pdf')
    fig.savefig(path + '/falsePositives.png')
    plt.close()
    
    fig = plt.figure(figsize=(20,7))
    ax = sns.lineplot(data=data, y="mismatches", x="frame-number", palette='pastel', linewidth=1)
    ax.axhline(data["mismatches"].mean(), linestyle='--', color='black')
    ax.set(xlabel='', ylabel='Mismatches')
    ax.yaxis.grid()
    fig.savefig(path + '/mismatches.pdf')
    fig.savefig(path + '/mismatches.png')
    plt.close()
    
    fig = plt.figure(figsize=(20,7))
    ax = sns.lineplot(data=data, y="matches", x="frame-number", palette='pastel', linewidth=1)
    ax.axhline(data["matches"].mean(), linestyle='--', color='black')
    ax.set(xlabel='', ylabel='Matches')
    ax.yaxis.grid()
    fig.savefig(path + '/matches.pdf')
    fig.savefig(path + '/matches.png')
    plt.close()
    
    fig = plt.figure(figsize=(20,7))
    ax = sns.lineplot(data=data, y="mota", x="frame-number", palette='pastel', linewidth=1)
    ax.axhline(data["mota"].mean(), linestyle='--', color='black')
    ax.set(xlabel='', ylabel='MOTA')
    ax.yaxis.grid()
    fig.savefig(path + '/mota.pdf')
    fig.savefig(path + '/mota.png')
    plt.close()
    
    fig = plt.figure(figsize=(20,7))
    ax = sns.lineplot(data=data, y="motp", x="frame-number", palette='pastel', linewidth=1)
    ax.axhline(data["motp"].mean(), linestyle='--', color='black')
    ax.set(xlabel='', ylabel='MOTP')
    ax.yaxis.grid()
    fig.savefig(path + '/motp.pdf')
    fig.savefig(path + '/motp.png')
    plt.close()
    
def overall_results(path, aggregation, detection, tracking, results, runs, best):
    size=1.5*runs
    fig = plt.figure(figsize=(size, 7))    
    ax = sns.boxplot(data=aggregation, x="run", y="processing-time", width=0.4, palette='pastel', showfliers=False, linewidth=1)
    ax.set(xlabel='RUN', ylabel='Processing Time', title="Aggregation " + f'Best Run: {best["aggregation"]}')
    ax.yaxis.grid()
    fig.savefig(path + '/aggregation.pdf')
    fig.savefig(path + '/aggregation.png')
    plt.close()
    
    fig = plt.figure(figsize=(size, 7))
    ax = sns.barplot(data=detection, x="run", y="processing-time", palette='pastel', linewidth=1)
    ax.set(xlabel='RUN', ylabel='Processing Time', title="Detection " + f'Best Run: {best["detection"]}')
    ax.yaxis.grid()
    fig.savefig(path + '/detection.pdf')
    fig.savefig(path + '/detection.png')
    plt.close()
    
    fig = plt.figure(figsize=(size, 7))
    ax = sns.boxplot(data=tracking, x="run", y="processing-time", width=0.4, palette='pastel', showfliers=False, linewidth=1)
    ax.set(xlabel='RUN', ylabel='Processing Time', title="Tracking" + f'Best Run: {best["tracking"]}')
    ax.yaxis.grid()
    fig.savefig(path + '/tracking.pdf')
    fig.savefig(path + '/tracking.png')
    plt.close()
        
    fig = plt.figure(figsize=(size, 7))
    ax = sns.barplot(data=results, x="run", y="mota", palette='pastel', linewidth=1)
    ax.set(xlabel='RUN', ylabel='MOTA', title="Accuracy " + f'Best Run: {best["mota"]}')
    ax.yaxis.grid()
    fig.savefig(path + '/mota.pdf')
    fig.savefig(path + '/mota.png')
    plt.close()
    
    fig = plt.figure(figsize=(size, 7))
    ax = sns.barplot(data=results, x="run", y="motp", palette='pastel', linewidth=1)
    ax.set(xlabel='RUN', ylabel='MOTP', title="Precision " + f'Best Run: {best["motp"]}')
    ax.yaxis.grid()
    fig.savefig(path + '/motp.pdf')
    fig.savefig(path + '/motp.png')
    plt.close()

    fig = plt.figure(figsize=(size, 7))
    ax = sns.barplot(data=results, x="run", y="miss_ratio", palette='pastel', linewidth=1)
    ax.set(xlabel='RUN', ylabel='Miss Ratio', title="Precision " + f'Best Run: {best["miss_ratio"]}')
    ax.yaxis.grid()
    fig.savefig(path + '/miss_ratio.pdf')
    fig.savefig(path + '/miss_ratio.png')
    plt.close()

ap = argparse.ArgumentParser()
ap.add_argument("--runs", type=int, 
                    required=True,
                    help="number of benchmark runs")
ap.add_argument("--path", type=str, 
                    required=True,
                    help="root folder for the evaluation")
ARGS = vars(ap.parse_args())

full_aggregation = pandas.DataFrame()
full_tracking = pandas.DataFrame()
full_detection = pandas.DataFrame()
full_pipeline_results = pandas.DataFrame()

min_aggregation_processing = sys.float_info.max
min_tracking_processing = sys.float_info.max
min_detection_processing = sys.float_info.max
max_mota = -sys.float_info.max
min_motp = sys.float_info.max
min_miss_ratio = sys.float_info.max

best = dict()

for i in range(1, ARGS["runs"]+1):
    path = "{0:s}/run{1:03d}".format(ARGS["path"], i)
    aggregation = pandas.read_csv(path + '/aggregation.csv')
    aggregation = aggregation[:1950]
    tmp = aggregation.mean()["processing-time"]
    if tmp < min_aggregation_processing:
        min_aggregation_processing = tmp
        best["aggregation"] = i
    aggregation["run"] = i

    tracking = pandas.read_csv(path + '/tracking.csv')
    tracking = tracking[:1950]
    tracking["run"] = i
    tmp = tracking.mean()["processing-time"]
    if tmp < min_tracking_processing:
        min_tracking_processing = tmp
        best["tracking"] = i

    detection = pandas.read_csv(path + '/detection.csv')
    detection = detection[:200]
    detection["run"] = i
    tmp = detection.mean()["processing-time"]
    if tmp < min_detection_processing:
        min_detection_processing = tmp
        best["detection"] = i

    pipeline_results = pandas.read_csv(path + '/pipeline-results.csv')
        
    with open(path + '/results.json') as file:
        results = json.loads(file.read())
            
    aggregation_plots(path, aggregation, results["matching"]["Aggregation"])
    tracking_plots(path, tracking, results["matching"]["Tracking"])
    detection_plots(path, detection, results["matching"]["Detection"])
    pipeline_result_plots(path, pipeline_results)

    mota = results["results"].get("MOT-mota", 0.0)
    motp = results["results"].get("MOT-motp", 0.0)
    miss_ratio = results["results"].get("MOT-miss-ratio", 0.0)
    tmp = pandas.DataFrame({'mota': [mota], 'motp': [motp], 'miss_ratio': [miss_ratio], 'run': [i]})
    if mota > max_mota and not mota == 0:
        max_mota = mota
        best["mota"] = i
    if motp < min_motp and not motp == 0:
        max_motp = motp
        best["motp"] = i
    if miss_ratio < min_miss_ratio and not miss_ratio == 0:
        min_miss_ratio = miss_ratio
        best["miss_ratio"] = i

    full_aggregation = pandas.concat([full_aggregation, aggregation])
    full_tracking = pandas.concat([full_tracking, tracking])
    full_detection = pandas.concat([full_detection, detection])
    full_pipeline_results = pandas.concat([full_pipeline_results, tmp])

overall_results(ARGS["path"], full_aggregation, full_detection, full_tracking, full_pipeline_results, ARGS["runs"], best)