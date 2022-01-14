import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
import os
import datetime
import collections
import sys

if (os.getcwd() == '/home/jovyan'):
    os.chdir('work/qualtrics_analysis/nov_2021')

type_of_advising = {"Career" : "Q3.2", "Academic" : "Q3.3", "pre-health": "Q3.4", "Peers": "Q3.5", "Karen" : "Q3.6"}
academic_advisors = [""]
question_string_map = dict()
topics = list()

def initial_cleanup(data : pd.DataFrame) -> pd.DataFrame:
    data.iloc[0:1, data.columns.get_loc("Q2"):].apply(create_dict)
    ## get rid of unnecessary header info
    data = data.iloc[2:]
    include_columns = ['IPAddress', 'StartDate','Finished', 'Q2', 'Q3', 'Q3.2', 'Q3.2_21_TEXT','Q3.3','Q3.3_21_TEXT', 'Q3.4', 'Q3.5', 'Q3.6', 'Q3.6_4_TEXT', 'Q4_1', 'Q4_2','Q4_3', 'Q4_4', 'Q4_5', 'Q4_6', 'Q14', 'Q5']
    data = data.filter(include_columns)
    data.dropna(axis=1)
    return data

def get_topics(x: str) -> list:
    splitString = x.split(",")
    topics = list()
    for ss in splitString:
        topics.append(ss)
    return topics

def get_flat_topics(se : pd.Series) -> list:
    """
    se: Pandas serie of "Q3"
    returns a list with all topics flattened.
    """
    se = se.dropna()
    topics = [get_topics(x) for x in se]
    flat_topics = [topic for tops in topics for topic in tops]
    return flat_topics

def get_topic_counter(flat_topics : list) -> collections.Counter:
    """
    flat_topics: list of all topics discussed in appointments, flattened
    returns a counter of the topics discussed in appointments.
    """
    counter = collections.Counter(tuple(flat_topics))
    return counter

def get_counter(se: pd.Series) -> collections.Counter:
    flat_topics = get_flat_topics(se)
    return get_topic_counter(flat_topics)

def create_dict(x : pd.Series) -> None :
    question_string_map[x.name] = x.values[0]

def new_initial_cleanup(data : pd.DataFrame) -> pd.DataFrame:
    ## get rid of unnecessary header info
    data = data.iloc[2:]
    # include_columns = ['IPAddress', 'StartDate','Finished', 'Q1', 'Q2', 'Q3', 'Q3.21_TEXT', 'Q4_1', 'Q4_2','Q4_3', 'Q4_4', 'Q4_5', 'Q4_6', 'Q5']
    # data = data.filter(include_columns)
    return data

def merge_questions_3(data : pd.DataFrame) -> pd.DataFrame:
    ## merge questions 3
    question_list = ["Q3_1", "Q3_2", "Q3_24", "Q3_3", "Q3_8", "Q3_9", "Q3_10", "Q3_11", "Q3_13", "Q3_14", "Q3_22", "Q3_25", "Q3_26", "Q3_27", "Q3_28", "Q3_29", "Q3_31", "Q3_21"]

    data["Q3"] = data[question_list].apply(lambda x : ",".join(x[x.notnull()]), axis=1)

    return data
    

    

def import_data(filename):
    return pd.read_csv(filename)

def main(month: str, year: str, write: bool, fName: str) :
    # health = import_data(f"{month}_{year}_health.csv")
    # career = import_data(f"{month}_{year}_career.csv")
    # academic = import_data(f"{month}_{year}_academic.csv")
    # peer_and_beck = import_data(f"{month}_{year}_peer.csv")


    # ## data clean up
    # health.rename(columns={"Q3.4":"Q3"}, inplace=True)
    # career.rename(columns={"Q3.2":"Q3"}, inplace=True)
    # academic.rename(columns={"Q3.3":"Q3"}, inplace=True)
    # peer_and_beck.rename(columns={"Q3.5":"Q3"}, inplace=True)

    # df_dict = {
    #     "health" : health,
    #     "career" : career,
    #     "academic": academic, 
    #     "peer_and_bec": peer_and_beck,
    # }

    # for key, val in df_dict.items():
    #     df_dict[key] = initial_cleanup(val)
    #     df_dict[key].head()

    ## merge old and new data
    new_data = import_data(fName)
    data = new_initial_cleanup(new_data)
    # print(data.head())

    data = merge_questions_3(data)

    print(data.head())

    df_dict = dict()
    df_dict['health'] = data[data['Q1'] == 'Pre-Health advising']
    df_dict['career'] = data[data['Q1'] == 'Career Advising']
    df_dict['academic'] = data[data['Q1'] == 'Academic Advising']
    df_dict['peer_and_bec'] = data[data['Q1'] == 'Peer advising']
    
    counters = dict()
    for key, df in df_dict.items():
        counters[key] = get_counter(df['Q3'])
    counters
    
    

    if write:
        for key, dic in counters.items():
            tm = datetime.datetime.now().time().strftime("%H%M")
            pd.DataFrame.from_dict(dic, orient="index").to_excel(f"{key}_{month}_{year}_{tm}.xlsx")

    return counters, df_dict

counters, df_dict = main("dec", "2021", True, sys.argv[1])