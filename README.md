# RequestRepeater

Basic script to make the same http calls n number of times. Made it because the postman repeater is a paid feature.

## Prerequisite
Go 1.21

## Usage
The script can be run using the following command
go run repeater.go

The script takes in certain flags as well for ease of use, they are

```-n``` (int): Number of requests to repeat, defaults to 1000.

```-url``` (string): URL to call, defaults to http://localhost:8090/api/1/rest/feed/run/task/snaplogic/projects/shared/%20new%20pipeline%200%20Task12.

```-method``` (string): HTTP method to use, defaults to POST.

```-token``` (string): Bearer token for authorization, no default value (required).

```-delay``` (int): Delay in milliseconds between each request, defaults to 10ms.
