# Reference Kala(https://github.com/ajvb/kala)

[![Build Status](https://github.com/lovego/kala/actions/workflows/go.yml/badge.svg)](https://github.com/lovego/kala/actions/workflows/go.yml)
[![Coverage Status](https://coveralls.io/repos/github/lovego/kala/badge.svg?branch=master&1)](https://coveralls.io/github/lovego/kala)
[![Go Report Card](https://goreportcard.com/badge/github.com/lovego/kala)](https://goreportcard.com/report/github.com/lovego/kala)
[![Documentation](https://pkg.go.dev/badge/github.com/lovego/kala)](https://pkg.go.dev/github.com/lovego/kala@v0.3.4)

**Supports job to be executed only once on multiple nodes.**

Kala is a simplistic, modern, and performant job scheduler written in Go.  Features:

- Single binary
- No dependencies
- JSON over HTTP API
- Job Stats
- Configurable Retries
- Scheduling with ISO 8601 Date and Interval notation
- Dependent Jobs
- Persistent with several database drivers
- Web UI

## Breakdown of schedule string. (ISO 8601 Notation)

Example `schedule` string:

```
R2/2017-06-04T19:25:16.828696-07:00/PT10S
```

This string can be split into three parts:

```
Number of times to repeat/Start Datetime/Interval Between Runs
```

#### Number of times to repeat

This is designated with a number, prefixed with an `R`. Leave out the number if it should repeat forever.

Examples:

* `R` - Will repeat forever
* `R1` - Will repeat once
* `R231` - Will repeat 231 times.

#### Start Datetime

This is the datetime for the first time the job should run.

Kala will return an error if the start datetime has already passed.

Examples:

* `2017-06-04T19:25:16`
* `2017-06-04T19:25:16.828696`
* `2017-06-04T19:25:16.828696-07:00`
* `2017-06-04T19:25:16-07:00`

*To Note: It is recommended to include a timezone within your schedule parameter.*

#### Interval Between Runs

This is defined by the [ISO8601 Interval Notation](https://en.wikipedia.org/wiki/ISO_8601#Time_intervals).

It starts with a `P`, then you can specify years, months, or days, then a `T`, preceded by hours, minutes, and seconds.

Lets break down a long interval: `P1Y2M10DT2H30M15S`

* `P` - Starts the notation
* `1Y` - One year
* `2M` - Two months
* `10D` - Ten days
* `T` - Starts the time second
* `2H` - Two hours
* `30M` - Thirty minutes
* `15S` - Fifteen seconds

Now, there is one alternative. You can optionally use just weeks. When you use the week operator, you only get that. An example of using the week operator for an interval of every two weeks is `P2W`.

Examples:

* `P1DT1M` - Interval of one day and one minute
* `P1W` - Interval of one week
* `PT1H` - Interval of one hour.

### More Information on ISO8601

* [Wikipedia's Article](https://en.wikipedia.org/wiki/ISO_8601)

## Overview of routes

| Task | Method | Route |
| --- | --- | --- |
|Creating a Job | POST | /api/v1/job/ |
|Getting a list of all Jobs | GET | /api/v1/job/ |
|Getting a Job | GET | /api/v1/job/{id}/ |
|Deleting a Job | DELETE | /api/v1/job/{id}/ |
|Deleting all Jobs | DELETE | /api/v1/job/all/ |
|Getting metrics about a certain Job | GET | /api/v1/job/stats/{id}/ |
|Starting a Job manually | POST | /api/v1/job/start/{id}/ |
|Disabling a Job | POST | /api/v1/job/disable/{id}/ |
|Enabling a Job | POST | /api/v1/job/enable/{id}/ |
|Getting app-level metrics | GET | /api/v1/stats/ |


## /job

This route accepts both a GET and a POST. Performing a GET request will return a list of all currently running jobs.
Performing a POST (with the correct JSON) will create a new Job.

Note: When creating a Job, the only fields that are required are the `Name` and the `Command` field. But, if you omit the `Schedule` field, the job will be ran immediately.

Example:
```bash
$ curl http://127.0.0.1:8000/api/v1/job/
{"jobs":{}}
$ curl http://127.0.0.1:8000/api/v1/job/ -d '{"epsilon": "PT5S", "command": "bash /home/ajvb/gocode/src/github.com/ajvb/kala/examples/example-kala-commands/example-command.sh", "name": "test_job", "schedule": "R2/2017-06-04T19:25:16.828696-07:00/PT10S"}'
{"id":"93b65499-b211-49ce-57e0-19e735cc5abd"}
$ curl http://127.0.0.1:8000/api/v1/job/
{
    "jobs":{
        "93b65499-b211-49ce-57e0-19e735cc5abd":{
            "name":"test_job",
            "id":"93b65499-b211-49ce-57e0-19e735cc5abd",
            "command":"bash /home/ajvb/gocode/src/github.com/ajvb/kala/examples/example-kala-commands/example-command.sh",
            "owner":"",
            "disabled":false,
            "dependent_jobs":null,
            "parent_jobs":null,
            "schedule":"R2/2017-06-04T19:25:16.828696-07:00/PT10S",
            "retries":0,
            "epsilon":"PT5S",
            "success_count":0,
            "last_success":"0001-01-01T00:00:00Z",
            "error_count":0,
            "last_error":"0001-01-01T00:00:00Z",
            "last_attempted_run":"0001-01-01T00:00:00Z",
            "next_run_at":"2017-06-04T19:25:16.828794572-07:00"
        }
    }
}
```

## /job/{id}

This route accepts both a GET and a DELETE, and is based off of the id of the Job. Performing a GET request will return a full JSON object describing the Job.
Performing a DELETE will delete the Job.

Example:
```bash
$ curl http://127.0.0.1:8000/api/v1/job/93b65499-b211-49ce-57e0-19e735cc5abd/
{"job":{"name":"test_job","id":"93b65499-b211-49ce-57e0-19e735cc5abd","command":"bash /home/ajvb/gocode/src/github.com/ajvb/kala/examples/example-kala-commands/example-command.sh","owner":"","disabled":false,"dependent_jobs":null,"parent_jobs":null,"schedule":"R2/2017-06-04T19:25:16.828696-07:00/PT10S","retries":0,"epsilon":"PT5S","success_count":0,"last_success":"0001-01-01T00:00:00Z","error_count":0,"last_error":"0001-01-01T00:00:00Z","last_attempted_run":"0001-01-01T00:00:00Z","next_run_at":"2017-06-04T19:25:16.828737931-07:00"}}
$ curl http://127.0.0.1:8000/api/v1/job/93b65499-b211-49ce-57e0-19e735cc5abd/ -X DELETE
$ curl http://127.0.0.1:8000/api/v1/job/93b65499-b211-49ce-57e0-19e735cc5abd/
```

## /job/stats/{id}

Example:
```bash
$ curl http://127.0.0.1:8000/api/v1/job/stats/5d5be920-c716-4c99-60e1-055cad95b40f/
{"job_stats":[{"JobId":"5d5be920-c716-4c99-60e1-055cad95b40f","RanAt":"2017-06-03T20:01:53.232919459-07:00","NumberOfRetries":0,"Success":true,"ExecutionDuration":4529133}]}
```

## /job/start/{id}

Example:
```bash
$ curl http://127.0.0.1:8000/api/v1/job/start/5d5be920-c716-4c99-60e1-055cad95b40f/ -X POST
```

## /job/disable/{id}

Example:
```bash
$ curl http://127.0.0.1:8000/api/v1/job/disable/5d5be920-c716-4c99-60e1-055cad95b40f/ -X POST
```

## /job/enable/{id}

Example:
```bash
$ curl http://127.0.0.1:8000/api/v1/job/enable/5d5be920-c716-4c99-60e1-055cad95b40f/ -X POST
```

## /stats

Example:
```bash
$ curl http://127.0.0.1:8000/api/v1/stats/
{"Stats":{"ActiveJobs":2,"DisabledJobs":0,"Jobs":2,"ErrorCount":0,"SuccessCount":0,"NextRunAt":"2017-06-04T19:25:16.82873873-07:00","LastAttemptedRun":"0001-01-01T00:00:00Z","CreatedAt":"2017-06-03T19:58:21.433668791-07:00"}}
```

## Debugging Jobs

There is a command within Kala called `run` which will immediately run a command as Kala would run it live, and then gives you a response on whether it was successful or not. Allows for easier and quicker debugging of commands.

```bash
$ kala run "ruby /home/user/ruby/my_ruby_script.rb"
Command Succeeded!
$ kala run "ruby /home/user/other_dir/broken_script.rb"
FATA[0000] Command Failed with err: exit status 1
```

## Dependent Jobs

### How to add a dependent job

Check out this [example for how to add dependent jobs](https://github.com/ajvb/kala/blob/master/examples/python/example_dependent_jobs.py) within a python script.

### Notes on Dependent Jobs

* Dependent jobs follow a rule of First In First Out
* A child will always have to wait until a parent job finishes before it runs
* A child will not run if its parent job does not.
* If a child job is disabled, it's parent job will still run, but it will not.
* If a child job is deleted, it's parent job will continue to stay around.
* If a parent job is deleted, unless its child jobs have another parent, they will be deleted as well.
