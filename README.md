# HarbourBridge: Spanner Evaluation and Migration

[![cloudspannerecosystem](https://circleci.com/gh/cloudspannerecosystem/harbourbridge.svg?style=svg)](https://circleci.com/gh/cloudspannerecosystem/harbourbridge)

HarbourBridge is a stand-alone open source tool for Cloud Spanner evaluation and
migration, using data from an existing PostgreSQL, MySQL, SQL Server, Oracle or DynamoDB database.
The tool ingests schema and data from either a pg_dump/mysqldump file or directly
from the source database, and supports both schema and data migration. For schema
migration, HarbourBridge automatically builds a Spanner schema from the schema
of the source database. This schema can be customized using the HarbourBridge schema assistant and
a new Spanner database is created using the Spanner schema built.

For more details on schema customization and use of the schema assistant, see
[web/README](webv2/README.md). The rest of this README describes the command-line
capabilities of HarbourBridge.

HarbourBridge is designed to simplify Spanner evaluation and migration.
Certain features of relational databases, especially those that don't
map directly to Spanner features, are ignored, e.g. stored functions and
procedures, and sequences. Types such as integers, floats, char/text, bools,
timestamps, and (some) array types, map fairly directly to Spanner, but many
other types do not and instead are mapped to Spanner's `STRING(MAX)`.
In the case of DynamoDB, the schema is inferred based on a certain
amount of sampled data.

View HarbourBridge as a way to get up and running fast, so you can focus on
critical things like tuning performance and getting the most out of
Spanner. Expect that you'll need to tweak and enhance what HarbourBridge
produces.

## Data Migration

HarbourBridge supports two types of data migrations:

* Minimal Downtime migration - A minimal downtime migration consists of two components, migration of existing data from the database and the stream of changes (writes and updates) that are made to the source database during migration, referred to as change database capture (CDC). Using HarbourBridge, the entire process where Datastream reads data from the source database and writes to a GCS bucket and data flow reads data from GCS bucket and writes to spanner database can be orchestrated using a unified interface. Performing schema changes on the source database during the migration is not supported. This is the suggested mode of migration for most databases.

  Please note that in order to perform minimal downtime migration for **PostgreSQL** database a user needs to create a publication and replication slot as mentioned [here](https://cloud.google.com/datastream/docs/configure-your-source-postgresql-database#selfhostedpostgresql)

* Bulk Migration -  HarbourBridge reads data from source database and writes it to the database created in Cloud Spanner. Changes which happen to the source database during the bulk migration may or may not be written to Spanner. To achieve consistent version of data, stop writes on the source while migration is in progress, or use a read replica. Performing schema changes on the source database during the migration is not supported. While there is no technical limit on the size of the database, it is recommended for migrating moderate-size datasets to Spanner(up to about 100GB).

For some quick starter examples on how to run HarbourBridge, take a look at
[Quickstart Guide](#quickstart-guide).

HarbourBridge automatically determines the cloud project to use, and generates
a new Spanner database name.
Command-line flags can be used to explicitly set the Spanner instance or
database name. See [Command line flags](#command-line-flags).

**WARNING: Please check that permissions for the Spanner instance used by
HarbourBridge are appropriate. Spanner manages access control at the database
level, and the database created by HarbourBridge will inherit default
permissions from the instance. All data written by HarbourBridge is visible to
anyone who can access the created database.**

As it processes the data, HarbourBridge reports on progress, provides
stats on the schema and data conversion steps, and an overall assessment of the
quality of the conversion. It also generates a schema file, report file and
a session file (and a bad-data file if data was dropped). See
[Files Generated by HarbourBridge](#files-generated-by-harbourbridge). Details
of how source database's schema is mapped to Spanner can be found in the
[Schema Conversion](#schema-conversion) section.

This tool is part of the Cloud Spanner Ecosystem, a community contributed and
supported open source repository. Please [report
issues](https://github.com/cloudspannerecosystem/harbourbridge/issues) and send
pull requests. See the [HarbourBridge
Whitepaper](https://github.com/cloudspannerecosystem/harbourbridge/blob/master/whitepaper.md)
for a discussion of our plans for the tool.

Note that the HarbourBridge tool is not an officially supported Google product
and is not officially supported as part of the Cloud Spanner product.

## Quickstart Guide

### Before you begin

Complete the steps described in
[Set up](https://cloud.google.com/spanner/docs/getting-started/set-up), which
covers creating and setting a default Google Cloud project, enabling billing,
enabling the Cloud Spanner API, and setting up OAuth 2.0 to get authentication
credentials to use the Cloud Spanner API.

In particular, ensure that you run

```sh
gcloud auth application-default login
```

to set up your local development environment with authentication credentials.

Set the GCLOUD_PROJECT environment variable to your Google Cloud project ID:

```sh
export GCLOUD_PROJECT=my-project-id
```

If you do not already have a Cloud Spanner instance, or you want to use a
separate instance specifically for running HarbourBridge, then create a Cloud
Spanner instance by following the "Create an instance" instructions on the
[Quickstart using the console](https://cloud.google.com/spanner/docs/quickstart-console)
guide. HarbourBridge will create a database for you, but it will not create a
Spanner instance.

Install Go ([download](https://golang.org/doc/install)) on your development
machine if it is not already installed, configure the GOPATH environment
variable if it is not already configured, and
[test your installation](https://golang.org/doc/install#testing).

### Installing HarbourBridge

You can make a copy of the HarbourBridge codebase from the github repository
and use "go run".

```sh
git clone https://github.com/cloudspannerecosystem/harbourbridge
cd harbourbridge
go run github.com/cloudspannerecosystem/harbourbridge help
```

Examples below assume that `harbourbridge` alias is set as following

```sh
alias harbourbridge="go run github.com/cloudspannerecosystem/harbourbridge"
```

This workflow also allows you to modify or customize the HarbourBridge codebase.

### Running HarbourBridge

To use the tool on a PostgreSQL database called mydb, run

```sh
pg_dump mydb > mydb.pg_dump
harbourbridge schema-and-data -source=postgresql < mydb.pg_dump
```

To use the tool on a MySQL database called mydb, run

```sh
mysqldump mydb > mydb.mysqldump
harbourbridge schema-and-data -source=mysql < mydb.mysqldump
```

To use the tool on a DynamoDB database, run

```sh
harbourbridge schema-and-data -source=dynamodb
```

Note: HarbourBridge accepts pg_dump/mysqldump's standard plain-text format,
but not archive or custom formats. More details on HarbourBridge example usage
can be found here:

- [PostgreSQL example usage](sources/postgres/README.md#example-postgresql-usage)
- [MySQL example usage](sources/mysql/README.md#example-mysql-usage)
- [DynamoDB example usage](sources/dynamodb/README.md#example-dynamodb-usage)
- [CSV example usage](sources/csv/README.md#example-csv-usage)
- [SQL Server example usage](sources/sqlserver/README.md#example-sqlserver-usage)
- [Oracle DB example usage](sources/oracle/README.md#example-oracle-usage)

This command will use the cloud project specified by the `GCLOUD_PROJECT`
environment variable, automatically determine the Cloud Spanner instance
associated with this project, convert the source schema to a Spanner schema
(For MySQL/Postgres) or infer a schema from the DynamoDB instance, create a
new Cloud Spanner database with this schema, and finally, populate this new
database with the data from the source database.
If the project has multiple instances, then list of available instances
will be shown and you will have to pick one of the available instances and
set the `--instance` flag in target-profile. The new Cloud Spanner database
will have a name of the form `{SOURCE}_{DATE}_{RANDOM}`, where`{SOURCE}` is the
value of the source flag,`{DATE}`is today's date, and`{RANDOM}` is a random
suffix for uniqueness.

See the [Troubleshooting Guide](#troubleshooting-guide) for help on debugging
issues.

HarbourBridge also [generates several files](#files-generated-by-harbourbridge)
when it runs: a schema file, a report file (with detailed analysis of the
conversion), a session file and a bad data file (if any data was dropped).

### Setting up the emulator

To run migrations against a local instance without having to connect to Cloud
spanner each time follow the following steps:

- **Start the emulator:**
    ```sh
    gcloud emulators spanner start
    ```
- **Set the SPANNER_EMULATOR_HOST:**
    ```sh
    export SPANNER_EMULATOR_HOST=localhost:9010
    ```

### Sample Dump Files

If you don't have ready access to a PostgreSQL or MySQL database, some example
dump files can be found [here](examples). The files
[cart.pg_dump](examples/cart.pg_dump) and
[cart.mysqldump](examples/cart.mysqldump) contain pg_dump and mysqldump output
for a very basic shopping cart application (just two tables, one for products
and one for user carts). The files [singers.pg_dump](examples/singers.pg_dump)
and [singers.mysqldump](examples/singers.mysqldump) contain pg_dump and
mysqldump output for a version of the [Cloud Spanner
singers](https://cloud.google.com/spanner/docs/schema-and-data-model#creating_a_table)
example. To use HarbourBridge on cart.pg_dump, download the file locally and run

```sh
harbourbridge schema -source=postgresql < cart.pg_dump
```

### Verifying Results

Once the tool has completed, you can verify the new database and its content
using the Google Cloud Console. Go to the [Cloud Spanner Instances
page](https://console.cloud.google.com/spanner/instances), select your Spanner
instance, and then find the database created by HarbourBridge and select
it. This will list the tables created by HarbourBridge. Select a table, and take
a look at its schema and data. Next, go to the query page, and try
some SQL statements. For example

```sql
SELECT COUNT(*) from mytable
```

to check the number of rows in table `mytable`.

### Next Steps

The tables created by HarbourBridge provide a starting point for evaluation of
Spanner. While they preserve much of the core structure of your PostgreSQL/MySQL
schema and data, many key features have been dropped, including functions,
sequences, procedures,triggers, and views. For DynamoDB, the conversion from
schemaless to schema is focused on the use-case where customers use DynamoDB in
a consistent, structured way with a fairly well defined set of columns and types.

As a result, the out-of-the-box performance you get from these tables could be
slower than what you get from PostgreSQL/MySQL/DynamoDB.

To improve performance, also consider using [Interleaved
Tables](https://cloud.google.com/spanner/docs/schema-and-data-model#creating-interleaved-tables)
to tune performance.

View HarbourBridge as a base set of functionality for Spanner evalution that can
be readily expanded. Consider forking and modifying the codebase to add the
functionality you need. Please [file
issues](https://github.com/cloudspannerecosystem/harbourbridge/issues) and send
PRs for fixes and new functionality. See our backlog of [open
issues](https://github.com/cloudspannerecosystem/harbourbridge/issues). Our
plans and aspirations for developing HarbourBridge further are outlined in the
[HarbourBridge
Whitepaper](https://github.com/cloudspannerecosystem/harbourbridge/blob/master/whitepaper.md).

You can also change the way HarbourBridge behaves by directly editing the
pg_dump/mysqldump output. For example, suppose you want to try out different
primary keys for a table. First run pg_dump/mysqldump and save the output to
a file. Then modify (or add) the relevant
`ALTER TABLE ... ADD CONSTRAINT ... PRIMARY KEY ...` statement in the
pg_dump/mysqldump output file so that the primary keys match what you need.
Then run HarbourBridge on the modified pg_dump/mysqldump output.

## Files Generated by HarbourBridge

HarbourBridge generates several files as it runs:

- Schema file (ending in `schema.txt`): contains the generated Spanner
  schema, interspersed with comments that cross-reference to the relevant
  PostgreSQL/MySQL schema definitions.

- Session file (ending in `session.json`): contains all schema and data
  conversion state endcoded as JSON. It is basically a snapshot of the session.

- Report file (ending in `report.txt`): contains a detailed analysis of the
  PostgreSQL/MySQL to Spanner migration, including table-by-table stats and an
  analysis of PostgreSQL/MySQL types that don't cleanly map onto Spanner types.
  Note that PostgreSQL/MySQL types that don't have a corresponding Spanner type
  are mapped to STRING(MAX).

- Bad data file (ending in `dropped.txt`): contains details of data
  that could not be converted and written to Spanner, including sample
  bad-data rows. If there is no bad-data, this file is not written (and we
  delete any existing file with the same name from a previous run).

By default, these files are prefixed by the name of the Spanner database (with a
dot separator). The file prefix can be overridden using the `-prefix`
[option](#options).

## HarbourBridge CLI (command line interface)

HarbourBridge CLI follows [subcommands](https://github.com/google/subcommands)
structure with the the following general syntax:

```sh
harbourbridge <subcommand> flags
```

### Getting Help

The command `harbourbridge help` displays the available subcommands and the important global flags.

```text
    commands   list all subcommand names
    help   describe subcommands and their syntax
```

To get help on individual subcommands, use

```sh
    harbourbridge help <subcommand>
```

This will print the usage pattern, a few examples, and a list of all available subcommand flags.

### Subcommands

#### harbourbridge `schema`

This subcommand can be used to perform schema conversion and report on the quality of the conversion. The generated schema mapping file (session.json) can be then further edited using the HarbourBridge web UI to make custom edits to the destination schema. This session file
is then passed to the data subcommand to perform data migration while honoring the defined
schema mapping. HarbourBridge also generates Spanner schema which users can modify manually and use directly as well.

#### harbourbridge `data`

This subcommand will perform data migration and report on the quality of the same. Rows which could not be migrated are reported in
dropped.txt file. This subcommand requires users to pass the session file (which contains schema mapping) generated by either the `schema` subcommand or web UI.

#### harbourbridge `schema-and-data`

This subcommand will generate a schema as well as perform data migration and report on the quality of both schema migration and data migration. This subcommand can be used to do a quick evaluation for the migration and get started quickly on Spanner.

#### harbourbridge `web`

This subcommand will run the Harbourbridge UI locally. The UI can be used to perform assisted schema and data migration.

### Command line flags

This section describes the flags common across all the subcommands. For flags
specific to a give subcommand run `harbourbridge help <subcommand>`.

`-source` Required flag. Specifies the source source. Supported sources 
are _'postgres'_, _'mysql'_, _'dynamodb'_ and _'csv'_(only in data mode).

`-target` Optional flag. Specifies the target database. Defaults to _'spanner'_
, which is the only supported target database today.

`-prefix` Specifies a file prefix for the report, schema, and bad-data files
written by the tool. If no file prefix is specified, the name of the Spanner
database (plus a '.') is used.

`-v` or `-verbose` Specifies verbose mode. This will cause HarbourBridge to
output detailed messages about the conversion.

`-skip-foreign-keys` Controls whether we add foreign key constraints after
data migration is complete. This flag cannot be used with schema-only mode,
and does not affect the generation of foreign key statements during schema
processing i.e. foreign key constraints will still appear in the generated
Spanner DDL files.

`-session` Specifies a session file that contains all schema and data
conversion state endcoded as JSON.

`-source-profile` Specifies detailed parameters for the source database such as connection parameters. See [Source Profile](#source-profile) for details.

`-target-profile` Specifies detailed parameters for the target database. See [Target Profile](#target-profile) for details.

`-dry-run` Controls whether we run the migration in dry run mode or not. Using this mode generates session file, schema and report for schema and/or data conversion without actually creating the Spanner database.

### Source Profile

HarbourBridge accepts the following params for --source-profile,
specified as "key1=value1,key2=value,..." pairs:

`file` Specifies the full path of the file to use for reading source database
schema and/or data. This param is optional, and the file can also be piped to
stdin, if available locally.

If the file is located in Google Cloud Storage (GCS), you can use the
following format: `file=gs://{bucket_name}/{path/to/file}`. Please ensure you
have read pemissions to the GCS bucket you would like to use.

`format` Specifies the format of the file. This param is also optional, and
defaults to `dump`. This may be extended in future to support other formats
such as `csv`, `avro` etc.

`host` Specifies the host name for the source database.
If not specified in case of direct connection to the source database, HarbourBridge
fetches it from the environment variables([Example usage](#21-generating-pgdump-file)).

`user` Specifies the user for the source database.
If not specified in case of direct connection to the source database, HarbourBridge
fetches it from the environment variables([Example usage](#21-generating-pgdump-file)).

`dbName` Specifies the name of the source database.
If not specified in case of direct connection to the source database, HarbourBridge
fetches it from the environment variables([Example usage](#21-generating-pgdump-file)).

`port` Specifies the port for the source database.
If not specified in case of direct connection to the source database, HarbourBridge
fetches it from the environment variables([Example usage](#21-generating-pgdump-file)).

`password` Specifies the password for the source database.
If not specified in case of direct connection to the source database, HarbourBridge
fetches it from the environment variables([Example usage](#21-generating-pgdump-file)).

`streamingCfg` Optional flag. Specifies the file path for streaming config.
Please note that streaming migration is only supported for MySQL, Oracle and PostgreSQL databases currently.

### Target Profile

HarbourBridge accepts the following options for --target-profile,
specified as "key1=value1,key2=value,..." pairs:

`dbName` Specifies the name of the Spanner database to create. This must be a
new database. If dbName is not specified, HarbourBridge creates a new unique
dbName.

`instance` Specifies the Spanner instance to use. The new database will be
created in this instance. If not specified, the tool automatically determines an
appropriate instance using gcloud.

`dialect` Specifies the dialect of Spanner database. By default, Spanner
databases are created with GoogleSQL dialect. You can override the same by
setting `dialect=PostgreSQL` in the `-target-profile`. Learn more about support
for PostgreSQL dialect in Cloud Spanner at
<https://cloud.google.com/spanner/docs/postgresql-interface>.

## Schema Conversion

Details on HarbourBridge schema conversion can be found here:

- [PostgreSQL schema conversion](sources/postgres/README.md#schema-conversion)
- [MySQL schema conversion](sources/mysql/README.md#schema-conversion)
- [DynamoDB schema conversion](sources/dynamodb/README.md#schema-conversion)
- [SQL Server schema conversion](sources/sqlserver/README.md#schema-conversion)
- [Oracle DB schema conversion](sources/oracle/README.md#schema-conversion)

## Data Migration

### Data Conversion

HarbourBridge converts data from the source to Spanner data based on
the Spanner schema it constructs. Conversion for most data types is fairly
straightforward, but several types deserve discussion. Details on HarbourBridge
data conversion can be found here:

- [PostgreSQL data conversion](sources/postgres/README.md#data-conversion)
- [MySQL data conversion](sources/mysql/README.md#data-conversion)
- [DynamoDB data conversion](sources/dynamodb/README.md#data-conversion)
- [CSV data conversion](sources/csv/README.md#example-csv-usage)
- [SQL Server data conversion](sources/sqlserver/README.md#data-conversion)

### Data Migration Recommendations
- While using direct connect, it is recommended to use a secondary/read replica
 to ensure consistency and that and avoid impact from the load on the primary.

## Troubleshooting Guide

The following steps can help diagnose common issues encountered while running
HarbourBridge.

### 1. Verify source profile configuration

First, check that the source profile is correctly configured to connect to your
database. Source profile configuration varies depending on the database.

#### 1.1 Direct access to PostgreSQL

See [Directly connecting to a PostgreSQL database](sources/postgres/README.md#directly-connecting-to-a-postgresql-database) for troubleshooting direct access to PostgreSQL.

#### 1.2 Direct access to MySQL

See [Directly connecting to a MySQL database](sources/mysql/README.md#directly-connecting-to-a-mysql-database) for troubleshooting direct access to MySQL.

#### 1.3 Direct access to DynamoDB

See [DynamoDB example usage](sources/dynamodb/README.md#example-dynamodb-usage)
for troubleshooting direct access to DynamoDB.
#### 1.4 Direct access to SQL Server

See [SQL Server example usage](sources/sqlserver/README.md#example-sqlserver-usage)
for troubleshooting direct access to SQL Server.

#### 1.5 Direct access to Oracle DB

See [Oracle DB example usage](sources/oracle/README.md#example-oracle-usage)
for troubleshooting direct access to Oracle.


### 2. Generating dump files

#### 2.1 Generating pg_dump file

If you are using pg_dump , check that pg_dump is
correctly configured to connect to your PostgreSQL
database. Note that pg_dump uses the same options as psql to connect to your
database. See the [psql](https://www.postgresql.org/docs/9.3/app-psql.html) and
[pg_dump](https://www.postgresql.org/docs/9.3/app-pgdump.html) documentation.

Access to a PostgreSQL database is typically configured using the
_PGHOST_, _PGPORT_, _PGUSER_, _PGDATABASE_ environment variables,
which are standard across PostgreSQL utilities.

It is also possible to configure access via pg_dump's command-line options
`--host`, `--port`, and `--username`.

#### 2.2 mysqldump

If you are using mysqldump, check that mysqldump is correctly configured to
connect to your MySQL database via the command-line options `--host`, `--port`,
and `--user`. Note that mysqldump uses the same options as mysql to connect to
your database. See the
[mysql](https://dev.mysql.com/doc/refman/8.0/en/mysql-commands.html) and
[mysqldump](https://dev.mysql.com/doc/refman/8.0/en/mysqldump.html)
documentation.

Next, verify that pg_dump/mysqldump is generating plain-text output. If your
database is small, try running

```sh
{ pg_dump/mysqldump } > file
```

and look at the output file. It should be a plain-text file containing SQL
commands. If your database is large, consider just dumping the schema via the
`--schema-only` for pg_dump and `--no-data` for mysqldump command-line option.

pg_dump/mysqldump can export data in a variety of formats, but HarbourBridge
only accepts `plain` format (aka plain-text). See the
[pg_dump documentation](https://www.postgresql.org/docs/9.3/app-pgdump.html) and
[mysqldump documentation](https://dev.mysql.com/doc/refman/8.0/en/mysqldump.html)
for details about formats.

### 3. Debugging HarbourBridge

The HarbourBridge tool can fail for a number of reasons.

#### 3.1 No space left on device

HarbourBridge needs to read the pg_dump/mysqldump output twice, once to build
a schema and once for data ingestion. When pg_dump/mysqldump output is directly
piped to HarbourBridge, `stdin` is not seekable, and so we write the output to
a temporary file. That temporary file is created via Go's ioutil.TempFile.
On many systems, this creates a file in `/tmp`, which is sometimes configured
with minimal space. A simple workaround is to separately run pg_dump/mysqldump
and write its output to a file in a directory with sufficient space. For example,
if the current working directory has space, then:

```sh
{ pg_dump/mysqldump } > tmpfile
harbourbridge < tmpfile
```

Make sure you cleanup the tmpfile after HarbourBridge has been run. Another
option is to set the location of Go's TempFile e.g. by setting the `TMPDIR`
environment variable.

#### 3.2 Unparsable dump output

HarbourBridge uses the [pg_query_go](https://github.com/pganalyze/pg_query_go)
library for parsing pg_dump and [pingcap parser](https://github.com/pingcap/parser)
for parsing mysqldump. It is possible that the pg_dump/mysqldump output is
corrupted or uses features that aren't parseable. Parsing errors should
generate an error message of the form `Error parsing last 54321 line(s) of input`.

#### 3.2 Credentials problems

HarbourBridge uses standard Google Cloud credential mechanisms for accessing
Cloud Spanner. If this is mis-configured, you may see errors containing
"unauthenticated", or "cannot fetch token", or "could not find default
credentials". You might need to run `gcloud auth application-default login`.
See the [Before you begin](#before-you-begin) section for details.

#### 3.4 Can't create database

In this case, the error message printed by the tool should help identify the
cause. It could be an API permissions issue. For example, the Cloud Spanner API
may not be appropriately configured. See [Before you begin](#before-you-begin)
section for details. Alternatively, you have have hit the limit on the number of
databases per instances (currently 100). This can occur if you re-run the
HarbourBridge tool many times, since each run creates a new database. In this
case you'll need to [delete some
databases](https://cloud.google.com/spanner/docs/getting-started/go/#delete_the_database).

### 4. Database-Specific Issues

The schema, report, and bad-data files [generated by
HarbourBridge](#files-generated-by-harbourbridge) contain detailed information
about the schema and data conversion process, including issues and problems
encountered.

### 5. Reporting Issues

## Known Issues
Please refer to the [issues section](https://github.com/cloudspannerecosystem/harbourbridge/issues)
 on Github for a full list of known issues.

### Schema Conversion

- Loading dump files from SQL Server, Oracle and DynamoDB is not supported
- Schema Only Mode does not create foreign keys
- Migration of check constraints, functions and views is not supported
- Schema recommendations are based on static analysis of the schema only
- PG Spanner dialect support is limited, and is not currently available on the UI

### Minimal Downtime Data Migrations

- Minimal downtime migrations for SQL Server and DynamoDB are not supported
- Requires a direct connection to the database to run and hence will not be
 available while reading from Dump files.
- Expected downtime will be in the order of a few minutes while the pipeline gets
 flushed.
- This flow depends on Datastream, and all the [constraints of Datastream](https://cloud.google.com/datastream/docs/faq#behavior-and-limitations)
 apply to these migrations
- Migration from sharded databases is not natively supported
- Edits to primary keys and unique indexes are supported, but the user will 
need to ensure that the new primary key/unique indexes retain uniqueness in
the data. This is not verified during updation of the keys
- When the Spanner table PKs are different from the source keys, updates on the spanner PK columns can potentially lead to data inconsistencies. The updates can be potentially treated as a new insert or update some different row
- Interleaved rows and rows with foreign key constraints are retried 500 times.
 Exhaustion of retries results in these rows being pushed into a dead letter queue.
- Conversion to Spanner ARRAY type is currently not supported
- MySQL types BIT and TIME are not converted correctly


