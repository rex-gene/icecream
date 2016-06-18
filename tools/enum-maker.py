import re
import getopt
import sys
import os

opts, args = getopt.getopt(sys.argv[1:], "", ["skip-line=", "config-file-path=", "temp-file-path=" , "write-file-path=" , "help"])

def usage() :
    print sys.argv[0] + " <options>"
    print "version:1.1"
    print "option:"
    print "\t--skip-line=<skip line count>"
    print "\t--config-file-path=<config file path>"
    print "\t--temp-file-path=<template file path>"
    print "\t--write-file-path=<output file path>"
    print "\t--help print usage"

def makeFile(skip_line, config_file_path, temp_file_path, write_file_path):
    if not os.path.exists(temp_file_path):
        print "[-]temp file not exists"
        return False

    getLoopZoneRegex = re.compile("\{#begin\}([\s\S]*)\{#end\}", re.M)

    file_paths = config_file_path.split("#")
    result = ""
    for file_path in file_paths:
        if not os.path.exists(file_path):
            print "[-]config file not exists:" + file_path
            return False

        config_file = open(file_path)
        line = 0

        temp_file = open(temp_file_path)
        temp_info = temp_file.read()
        temp_file.close()

        infoResult = getLoopZoneRegex.findall(temp_info)
        if len(infoResult) == 0:
            print "[-]temp file have not token"
            return False

        for data  in config_file.readlines():
            if line >= int(skip_line):
                feilds=data.split(',')        
                if len(feilds) < 2 :
                    print "[-]config file invalid"
                    return False


                id = feilds[0].strip()
                name = feilds[1].strip()

                info = infoResult[0]
                info = info.replace("{@name}", name)
                info = info.replace("{@id}", id)

                result = result + info


            line = line + 1

        config_file.close()

    writeData = getLoopZoneRegex.sub(result, temp_info)
    out_file = open(write_file_path, "w")
    out_file.write(writeData)
    out_file.close()


if __name__ == "__main__":
    skip_line = None
    config_file_path = None
    temp_file_path = None
    write_file_path = None

    for key, value in opts:
        if key == "--skip-line":
            skip_line = value
        elif key == "--config-file-path":
            config_file_path = value
        elif key == "--temp-file-path":
            temp_file_path = value
        elif key == "--write-file-path":
            write_file_path = value
        elif key == "--help":
            usage()
            sys.exit()

    if skip_line == None or config_file_path == None or temp_file_path == None or write_file_path == None:
        usage()
        sys.exit()

    makeFile(skip_line, config_file_path, temp_file_path, write_file_path)

