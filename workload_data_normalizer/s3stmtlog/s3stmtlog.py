import json

def stmt_sample_user(users):
    for u in users:
        return json.dumps(u)
    return None

# {"begin":1712001865,"end":1712001925,"schema_name":"gharchive_dev","digest":"0809f1517e952f95740cf954e2de761aaca88641787d17b1869cad3115ec627b","plan_digest":"","stmt_type":"Insert","normalized_sql":"insert into `stats_api_requests` ( `client_ip` , `client_origin` , `method` , `path` , query , error , `status_code` , duration , `is_dev` ) values ( ... , false , ... , false )","table_names":"gharchive_dev.stats_api_requests","is_internal":false,"sample_sql":"INSERT INTO stats_api_requests(client_ip, client_origin, method, path, query, error, status_code, duration, is_dev) VALUES ('109.242.111.22', 'https://ossinsight.io', 'GET', '/q/events-total', '{}', false, 200, 15.927533984184265, false)","charset":"utf8mb4","collation":"utf8mb4_unicode_ci","prev_sql":"","sample_plan":"pATYMAkyN18xCTAJMAlOL0EJMAl0aW1lOjE0OC4xwrVzLCBsb29wczoxLCBwcmVwYXJlOiA1Ni40wgEbKGluc2VydDo5MS43BSuIY29tbWl0X3R4bjoge3ByZXdyaXRlOjEuNjZtcywgZ2V0X2MNIhBzOjkwNR00DDozLjEFJRxzbG93ZXN0XxE+8FJfcnBjOiB7dG90YWw6IDAuMDAycywgcmVnaW9uX2lkOiA1MTM1NTY5NDQ2LCBzdG9yZTogdGlrdi1pM2VuM3hsYXJnZS0yNXYyMzA3MTEtcDEtdAEeCDAudAEHdiUAiHBlZXIudGlkYi1zZXJ2ZXJsZXNzLnN2YzoyMDE2MCwgfSwgDfIYcHJpbWFyeUatAAAz/q0A/q0ALq0ALTYIbnVtIc8lnRxfa2V5czo1LA0OXGJ5dGU6NDAzfQk3ODkgQnl0ZXMJTi9BCg==","sample_binary_plan":"rQTwQwqoBAoISW5zZXJ0XzE4AUABUgNOL0FaFnRpbWU6MTQ4LjHCtXMsIGxvb3BzOjFiIHByZXBhcmU6IDU2LjTCtXMsIGluAT2wOjkxLjfCtXNiygNjb21taXRfdHhuOiB7cHJld3JpdGU6MS42Nm1zLCBnZXRfESIQczo5MDUFRQkSEDozLjE2ASUcc2xvd2VzdF8RPvBSX3JwYzoge3RvdGFsOiAwLjAwMnMsIHJlZ2lvbl9pZDogNTEzNTU2OTQ0Niwgc3RvcmU6IHRpa3YtaTNlbjN4bGFyZ2UtMjV2MjMwNzExLXAxLXQBHggwLnQBB3YlAIhwZWVyLnRpZGItc2VydmVybGVzcy5zdmM6MjAxNjAsIH0sIAm+HF9wcmltYXJ5Rq0AADP+rQD+rQAurQAtNhhudW06MSwgJZ0YX2tleXM6NREOYGJ5dGU6NDAzfXCVBnj///////////8BGAE=","plan_hint":"","index_names":null,"exec_count":2,"sum_errors":0,"sum_warnings":0,"sum_latency":12406676,"max_latency":7499914,"min_latency":4906762,"sum_parse_latency":302399,"max_parse_latency":191248,"sum_compile_latency":398429,"max_compile_latency":199233,"sum_num_cop_tasks":0,"max_cop_process_time":0,"max_cop_process_address":"","max_cop_wait_time":0,"max_cop_wait_address":"","sum_process_time":0,"max_process_time":0,"sum_wait_time":0,"max_wait_time":0,"sum_backoff_time":0,"max_backoff_time":0,"sum_total_keys":0,"max_total_keys":0,"sum_processed_keys":0,"max_processed_keys":0,"sum_rocksdb_delete_skipped_count":0,"max_rocksdb_delete_skipped_count":0,"sum_rocksdb_key_skipped_count":0,"max_rocksdb_key_skipped_count":0,"sum_rocksdb_block_cache_hit_count":0,"max_rocksdb_block_cache_hit_count":0,"sum_rocksdb_block_read_count":0,"max_rocksdb_block_read_count":0,"sum_rocksdb_block_read_byte":0,"max_rocksdb_block_read_byte":0,"commit_count":2,"sum_get_commit_ts_time":1279273,"max_get_commit_ts_time":905023,"sum_prewrite_time":2901041,"max_prewrite_time":1664927,"sum_commit_time":4898745,"max_commit_time":3162118,"sum_local_latch_time":0,"max_local_latch_time":0,"sum_commit_backoff_time":0,"max_commit_backoff_time":0,"sum_resolve_lock_time":0,"max_resolve_lock_time":0,"sum_write_keys":10,"max_write_keys":5,"sum_write_size":782,"max_write_size":403,"sum_prewrite_region_num":2,"max_prewrite_region_num":1,"sum_txn_retry":0,"max_txn_retry":0,"sum_backoff_times":0,"backoff_types":{},"auth_users":{"3EDFHZJX5iSzvfr.gh_api":{}},"wru":24.11484375,"rru":0,"ru":24.11484375,"sum_mem":17950,"max_mem":8981,"sum_disk":0,"max_disk":0,"sum_affected_rows":2,"sum_kv_total":7480834,"sum_pd_total":2865631,"sum_backoff_total":0,"sum_write_sql_resp_total":0,"sum_result_rows":0,"max_result_rows":0,"min_result_rows":0,"prepared":false,"first_seen":"2024-04-01T20:04:31.540734338Z","last_seen":"2024-04-01T20:05:11.542344761Z","plan_in_cache":false,"plan_cache_hits":0,"plan_in_binding":false,"exec_retry_count":0,"exec_retry_time":0,"keyspace_name":"3EDFHZJX5iSzvfr","keyspace_id":3490,"serverless_tenant_id":"1372813089187041280","serverless_project_id":"1372813089206301327","serverless_cluster_id":"1379661944642684098"}
def s3stmt2sql(one_stmt_json):
    d = json.loads(one_stmt_json)
    cols = {
        "SUMMARY_BEGIN_TIME": "from_unixtime(%d)" % d['begin'],
        "SUMMARY_END_TIME": "from_unixtime(%d)" % d['end'],
        "DIGEST_TEXT": json.dumps(d['normalized_sql']),
        "SAMPLE_USER": stmt_sample_user(d['auth_users']),
        "QUERY_SAMPLE_TEXT": json.dumps(d['sample_sql']),
        "AVG_LATENCY": d['sum_latency'] / d['exec_count'],
        "AVG_PARSE_LATENCY": d['sum_parse_latency'] / d['exec_count'],
        "AVG_COMPILE_LATENCY": d['sum_compile_latency'] / d['exec_count'],
        "SUM_COP_TASK_NUM": d['sum_num_cop_tasks'],
        "AVG_PROCESS_TIME": d['sum_process_time'] / d['exec_count'],
        "AVG_WAIT_TIME": d['sum_wait_time'] / d['exec_count'],
        "AVG_BACKOFF_TIME": d['sum_wait_time'] / d['exec_count'],
        "AVG_TOTAL_KEYS": d['sum_total_keys'] / d['exec_count'],
        "AVG_PROCESSED_KEYS": d['sum_processed_keys'] / d['exec_count'],
        "AVG_ROCKSDB_DELETE_SKIPPED_COUNT": d['sum_rocksdb_delete_skipped_count'] / d['exec_count'],
        "AVG_ROCKSDB_KEY_SKIPPED_COUNT": d['sum_rocksdb_key_skipped_count'] / d['exec_count'],
        "AVG_ROCKSDB_BLOCK_CACHE_HIT_COUNT": d['sum_rocksdb_block_cache_hit_count'] / d['exec_count'],
        "AVG_ROCKSDB_BLOCK_READ_COUNT": d['sum_rocksdb_block_read_count'] / d['exec_count'],
        "AVG_ROCKSDB_BLOCK_READ_BYTE": d['sum_rocksdb_block_read_byte'] / d['exec_count'],
        "AVG_PREWRITE_TIME": d['sum_prewrite_time'] / d['exec_count'],
        "AVG_COMMIT_TIME": d['sum_commit_time'] / d['exec_count'],
        "AVG_GET_COMMIT_TS_TIME": d['sum_get_commit_ts_time'] / d['exec_count'],
        "AVG_COMMIT_BACKOFF_TIME": d['sum_commit_backoff_time'] / d['exec_count'],
        "AVG_RESOLVE_LOCK_TIME": d['sum_resolve_lock_time'] / d['exec_count'],
        "AVG_LOCAL_LATCH_WAIT_TIME": d['sum_local_latch_time'] / d['exec_count'],
        "MAX_LOCAL_LATCH_WAIT_TIME": d['max_local_latch_time'],
        "AVG_WRITE_KEYS": d['sum_write_keys'] / d['exec_count'],
        "AVG_WRITE_SIZE": d['sum_write_size'] / d['exec_count'],
        "AVG_PREWRITE_REGIONS": d['sum_prewrite_region_num'] / d['exec_count'],
        "MAX_PREWRITE_REGIONS": d['max_prewrite_region_num'],
        "AVG_TXN_RETRY": d['sum_txn_retry'],
        "AVG_MEM": d['sum_mem'] / d['exec_count'],
        "SUM_EXEC_RETRY": d['exec_retry_count'],
        "SUM_EXEC_RETRY_TIME": d['exec_retry_time'],
        "AVG_DISK": d['sum_disk'] / d['exec_count'],
        "AVG_KV_TIME": d['sum_kv_total'] / d['exec_count'],
        "AVG_PD_TIME": d['sum_pd_total'] / d['exec_count'],
        "AVG_BACKOFF_TOTAL_TIME": d['sum_backoff_total'] / d['exec_count'],
        "AVG_WRITE_SQL_RESP_TIME": d['sum_write_sql_resp_total'] / d['exec_count'],
        "AVG_RESULT_ROWS": d['sum_result_rows'] / d['exec_count'],
        "AVG_AFFECTED_ROWS": d['sum_affected_rows'] / d['exec_count'],

        # TODO
        "BACKOFF_TYPES": json.dumps(""),
        "PLAN_HINT": json.dumps(""),
        "PLAN": json.dumps(""),
        # "PREV_SAMPLE_TEXT": json.dumps(d['prev_sql']),
        # "BINARY_PLAN": json.dumps(d['sample_binary_plan']),
        # "RESOURCE_GROUP": json.dumps(""), # TODO,
        # "AVG_REQUEST_UNIT_READ": d['sum_request_unit_read'] / d['exec_count'],
        # "AVG_REQUEST_UNIT_WRITE": d['sum_request_unit_write'] / d['exec_count'],
        # "AVG_QUEUED_RC_TIME": d['sum_queued_rc_time'] / d['exec_count']
    }
    # "AVG_PD_TIME","AVG_BACKOFF_TOTAL_TIME","AVG_WRITE_SQL_RESP_TIME",
    str_cols = ["STMT_TYPE", "SCHEMA_NAME", "DIGEST", 
                "TABLE_NAMES", "INDEX_NAMES", 
                "MAX_COP_PROCESS_ADDRESS", "MAX_COP_WAIT_ADDRESS",
                "PLAN_DIGEST", "CHARSET", "COLLATION",
                "FIRST_SEEN","LAST_SEEN"]
    for col in str_cols:
        cols[col] = json.dumps(d[col.lower()])

    num_cols = ["EXEC_COUNT", "SUM_ERRORS", "SUM_WARNINGS",
                "SUM_LATENCY", "MAX_LATENCY", "MIN_LATENCY",
                "MAX_PARSE_LATENCY", "MAX_COMPILE_LATENCY", 
                "MAX_COP_PROCESS_TIME", "MAX_COP_WAIT_TIME", 
                "MAX_PROCESS_TIME", "MAX_WAIT_TIME",
                "MAX_BACKOFF_TIME", "MAX_TOTAL_KEYS", 
                "MAX_PROCESSED_KEYS", "MAX_ROCKSDB_DELETE_SKIPPED_COUNT",
                "MAX_ROCKSDB_KEY_SKIPPED_COUNT",
                "MAX_ROCKSDB_BLOCK_CACHE_HIT_COUNT",
                "MAX_ROCKSDB_BLOCK_READ_COUNT", "MAX_ROCKSDB_BLOCK_READ_BYTE",
                "MAX_PREWRITE_TIME","MAX_COMMIT_TIME",
                "MAX_GET_COMMIT_TS_TIME","MAX_COMMIT_BACKOFF_TIME",
                "MAX_RESOLVE_LOCK_TIME",
                "MAX_WRITE_KEYS","MAX_WRITE_SIZE",
                "MAX_TXN_RETRY",
                "SUM_BACKOFF_TIMES","MAX_MEM","MAX_DISK","MAX_RESULT_ROWS",
                "MIN_RESULT_ROWS","PREPARED",
                "PLAN_IN_CACHE","PLAN_CACHE_HITS","PLAN_IN_BINDING",
                # "MAX_REQUEST_UNIT_READ","MAX_REQUEST_UNIT_WRITE","MAX_QUEUED_RC_TIME",
                ]
    for col in num_cols:
        cols[col] = d[col.lower()]

    for col in cols:
        cols[col] = str(cols[col])

    sql_cols = []
    sql_vals = []
    for col in sorted(cols.keys()):
        sql_cols.append(col)
        sql_vals.append(cols[col])

    sql = "insert into statements_summary (%s) values (%s);" % (", ".join(sql_cols), ", ".join(sql_vals))
    return sql

f = "./s3stmtlog.log"
with open(f, 'r') as file:
    for line in file:
        line = line.strip()
        if line == "":
            continue
        print(s3stmt2sql(line))