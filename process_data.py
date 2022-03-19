import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
import os
import datetime
import collections
import sys
import sqlite3
from datetime import datetime

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

    data.loc[ : ,"Q3"] = data[question_list].apply(lambda x : ",".join(x[x.notnull()]), axis=1)

    return data


def filter_date(data: pd.DataFrame, month : str, year : str) -> pd.DataFrame:
    start = f'{year}-{month}-01'
    end = f'{year}-{month}-31'
    pd.to_datetime(data["StartDate"])
    data = data[(data["StartDate"] > start) & (data["StartDate"] < end)]

    return data
    

def import_data(filename):
    return pd.read_csv(filename)

def addYearAndMonth(data : pd.DataFrame) -> pd.DataFrame :
    data.loc[:, "StartDate"] = pd.to_datetime(data.StartDate)
    data["StartMonth"] = data.StartDate.dt.month
    data["StartYear"] = data.StartDate.dt.year
    return data
    
# add/append data if it already exists 

def main(fName: str, dbname: str):
    new_data = import_data(fName)
    new_data = new_initial_cleanup(new_data)

    new_data = merge_questions_3(new_data)

    ## filter data for month
    ## rather than filtering, we should group it together and add all the grouped months.
    # data = filter_date(data, month, year)
    new_data = addYearAndMonth(new_data)
    if not os.path.exists(dbname):
        con = sqlite3.connect(dbname) 
        cur = con.cursor()
        cur.execute('''CREATE TABLE records 
        (
            id integer not null primary key autoincrement, 
            department text not null, 
            month text not null, 
            year text not null,
            unique(month, year, department)
        );
        ''')
        cur.execute('''CREATE TABLE topic_counts (
            id integer not null primary key autoincrement,
            record_id integer not null,
            topic text not null, 
            count integer not null,
            FOREIGN KEY (record_id) REFERENCES records(id)
            unique(record_id, topic)
            );''')
    else :
        con = sqlite3.connect(dbname)
        cur = con.cursor()

    for idx, data in new_data.groupby(["StartYear", "StartMonth"]):
        month = str(data.StartMonth.iloc[0])
        year = str(data.StartYear.iloc[0])
        df_dict = dict()
        df_dict['health'] = data[data['Q1'] == 'Pre-Health advising']
        df_dict['career'] = data[data['Q1'] == 'Career Advising']
        df_dict['academic'] = data[data['Q1'] == 'Academic Advising']
        df_dict['peer'] = data[data['Q1'] == 'Peer advising']

        counters = dict()
        for key, df in df_dict.items():
            counters[key] = get_counter(df['Q3'])



        print(month)
        print(year)
        for dep, dic in counters.items():
            stmt = 'insert into records (department, month, year) values (?, ?, ?);'
            try:
                cur.execute(stmt, (dep, month, year))
                id = cur.lastrowid
            except:
                id = cur.execute('select id from records where department=? and month=? and year=?;', [dep, month, year]).fetchone()[0]
                
            for topic, val in dic.items():
                try:
                    id, count  = cur.execute("select id, count from topic_counts where record_id=? and topic=?;", [id, topic]).fetchone()
                    type(count)
                    type(val)
                    count += val
                    stmt = 'update topic_counts set count=? where id=?'
                    cur.execute(stmt, [count, id])
                except:
                    count = val
                    stmt = 'insert into topic_counts (record_id, topic, count) values (?, ?, ?);'
                    cur.execute(stmt, [id, topic, count])
                
            
    con.commit()
    con.close()



main(sys.argv[1], sys.argv[2])
