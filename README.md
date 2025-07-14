# Transaction Isolation levels in Postgres and Mysql

### There are 4 transaction isolation levels:
```
1. Read Uncommitted - 
2. Read Committed - 
3. Repeatable Read
4. Serializable

```

The phenomena which are prohibited at various levels are:
```
- dirty read
            A transaction reads data written by a ** concurrent uncommitted ** transaction.

- nonrepeatable read
            A transaction re-reads data it has previously read and finds that data has been modified by
            another trnsaction (that committed since the initial read).

- phantom read
            A transaction re-executes a query returning a set of rows that satisy a search conditon and finds that the set of rows satisfying the condition has changes due to another recently-committed transaction.

- serialization anomaly
            The result of successfully committing a group of transactions is incosistent with all possible orderings of running those transactions one at a time.

```
