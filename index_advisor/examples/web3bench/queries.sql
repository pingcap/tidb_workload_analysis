-- R1
Select to_address, from_address from transactions where hash = '0x1f415defb2729863fd8088727900d99b7df6f03d5e22e2105fc984cac3d0fb1c';

-- R21
Select * from transactions where to_address in ('0x70f0f4f40fed33420c1e4ceefa1eb482e044ba24', 
                                                '0x34662f274a42a17876926bc7b0ba541535e40e5f', 
                                               '0x7259c2a51a9b1f7e373dcd00898d26a44ffc2e7c');

-- R22
Select *  from transactions
where hash in ('0x1f415defb2729863fd8088727900d99b7df6f03d5e22e2105fc984cac3d0fb1c', 
               '0xbeff7a4cf341d10c6293a2ecfb255f39c21836bf8956c6877d0f2486794fd5b8', 
               '0x5dee984c63cc26037a81d0f2861565c4e0c21a87ebf165b331faec347d7a76a1', 
              '0xc7da1e3391e4b7769fffe8e6afc284175a6cbe5fd9b333d9c0585944a36118dd') and to_address <> from_address;

-- R31
SELECT * FROM token_transfers WHERE from_address = '0xfbb1b73c4f0bda4f67dca266ce6ef42f520fbb98' ORDER BY block_number DESC LIMIT 5;

-- R32
Select count(*) from token_transfers where token_address = '0x7a93f0d9f302c0818022f8dca6ee1eb0f1b50308';

-- R41
SELECT * FROM transactions
WHERE from_address = '0x31d118c5f75502b96ca21d3d0d3fb8d7b19fed24' OR to_address = '0x6364989a903f45798c7a292778285a83d0928608' 
ORDER BY block_timestamp DESC LIMIT 10;

-- R42
SELECT count(DISTINCT from_address) FROM transactions;

-- R43
SELECT
    sum(`value`) AS totalamount,
    count(`value`) AS transactioncount,
    from_address AS fromaddress
FROM transactions
WHERE to_address = '0xfeadad412ec5b5f62afe4b6f39a168eb5f098f41' AND 
      block_timestamp >= 1499637035 AND block_timestamp <= 1499639599 AND `value` > 1008000000000000
GROUP BY from_address
ORDER BY sum(value) DESC
    LIMIT 10;

-- R44
SELECT
    count(*) as count
FROM (SELECT *
    FROM token_transfers t
    WHERE from_address = '0xfbb1b73c4f0bda4f67dca266ce6ef42f520fbb98'
    UNION ALL
    SELECT t2.*
    FROM token_transfers t2
    INNER JOIN token_transfers t ON t2.from_address = t.to_address
    AND t.value < t2.value
    LIMIT 100) as temp;


-- R45
SELECT COUNT(DISTINCT block_receipts) as count
FROM (SELECT block_number AS block_receipts
    FROM receipts
    WHERE NOT EXISTS (
    SELECT block_number
    FROM transactions
    WHERE block_number = receipts.block_number)) as temp;
