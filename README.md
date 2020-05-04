# 'GO' - testing with mock
[![License MIT](https://img.shields.io/badge/license-MIT-lightgrey.svg?style=flat)](LICENSE)

## This example has been created from code from:
https://github.com/redhug1/go-ns/blob/feature/log-panics/mongo/

which came from:

https://github.com/ONSdigital/go-ns/tree/feature/update-rchttp-dataset-api-client

There are two folders with two different versions of the code.

The names of the folders indicates which example requires mongo.

Tested with Go 1.14


## Overview
The code demonstrates testing a `production` function called `shutdown()` and then 'replacing' it with a `mock'd` version to simulate some tests that can happen in production.

## Executing test
In whichever version folder of the code you choose, open both files in VSCODE and in the `mongo_test.go` file scroll down to:
```
func TestSuccessfulCloseMongoSession(t *testing.T) {
```
Where you should see Visual Studio Code offering two options just above this function (assuming you have delve debugger installed):
```
    run test|debug test
```
Click on one of these.

## Output from running the mongo version of code:
In the DEBUG CONSOLE of vscode you should see:
```
Exiting: cleanupTestData
'start' points to ==> mongo.graceful, {}
Doing Original: shutdown() ...
.Doing Original: shutdown() ...
.Doing Original: shutdown() ...
.
3 total assertions

Exiting: setUpTestData
NOTE: The test code will now switch to using the 'mock' function to handle simulated database issues ...
'start' points to ==> mongo.ungraceful, {}
Doing Test Harness: shutdown() ...
..Got key: early
Doing Test Harness: shutdown() ...
..
7 total assertions

ctx Err:context deadline exceeded
Exiting: cleanupTestData
Doing 2.5 second extra delay to allow session close's ...
Exiting: slowQueryMongo
Exiting: slowQueryMongo
PASS
```

# Study the code to see how it get's the `mock` version of `shutdown()` to be used ...

### Licence

Copyright ©‎ 2020, red

Released under MIT license, see [LICENSE](LICENSE.md) for details.