./index_advisor advise-offline --query-path='./examples/tpch_example1/queries' \
--schema-path='./examples/tpch_example1/schema.sql' \
--stats-path='./examples/tpch_example1/stats' \
--tidb-version='nightly' \
--output='./examples/tpch_example1/output' \
--max-num-indexes=5;

./index_advisor advise-offline --query-path='./examples/tpch_example2/queries.sql' \
--schema-path='./examples/tpch_example2/schema.sql' \
--stats-path='./examples/tpch_example2/stats' \
--tidb-version='nightly' \
--output='/tmp/index_advisor_output/tpch_example2' \
--max-num-indexes=5;

./index_advisor advise-offline --dir-path='./examples/job_fk' \
--tidb-version='nightly' \
--output='/tmp/index_advisor_output/job_fk' \
--max-num-indexes=5;

./index_advisor advise-offline --dir-path='./examples/job_no_fk' \
--tidb-version='nightly' \
--output='/tmp/index_advisor_output/job_no_fk' \
--max-num-indexes=5;

./index_advisor advise-offline --dir-path='./examples/web3bench' \
--tidb-version='nightly' \
--output='/tmp/index_advisor_output/web3bench' \
--max-num-indexes=5;

./index_advisor advise-offline --dir-path='./examples/tpcds' \
--tidb-version='nightly' \
--output='/tmp/index_advisor_output/tpcds' \
--max-num-indexes=5 \
--query-black-list='q5,q14,q18,q22,q27,q77,q80,q36,q86,q23,q51,q97,q67,q70,q78,q64,q41,q38,q81,q1,q30,q39,q54,q83,q31,q60,q33,q56,q58,q24,q57,q47,q95,q2,q59,q4,q11,q74';
