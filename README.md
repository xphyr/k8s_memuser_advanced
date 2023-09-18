# Memuser Advanced

This tool will use up memory to what ever amount you specify on the command line. This tool was written to test out kubernetes memory quotas, limits and autoscalers. The tool will allocate memory every time you connect to the /consumemem endpoint and it will clear up and deallocate memory when you hit the /clearmem endpoint. If you want to see current memory usage without changing memory usage you can see the status at /.

By default the application will run on port 8080.

## Usage

`./memuser -memory=<memory in mb> -maxmemory=<maximum memory to allocate> -fast=[true|false] -listenport=:<port number>`

## Parameters

 |  **Flag**   |                     **Description**                     | **Default Value** |
 | :---------: | :-----------------------------------------------------: | :---------------: |
 |   -memory   |      Consume this amount of memory each allocation      |       50 Mb       |
 | -maxmemory  | Dont use more than the specified amount of memory in Mb |      1000 Mb      |
 |    -fast    |                Ramp up memory usage fast                |       true        |
 | -listenport |           Port to listen to for http traffic            |       :8000       |


## Credits

code originaly from https://golangcode.com/print-the-current-memory-usage/

A change