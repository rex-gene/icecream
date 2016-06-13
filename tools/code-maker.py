import re
import sys
import getopt
import os

opts, args = getopt.getopt(sys.argv[1:], "", ["handler-template=", "protocol-template=", "init-template=", "handler-output=", "protocol-output=", "protocol-config=", "init-output=", "enum-output=", "enum-template=", "help"])

handler_tmp_file = None
protocol_tmp_file = None
init_tmp_file = None
handler_output = None
protocol_output = None
protocol_config = None
init_output = None
enum_output = None
enum_tmp_file = None

def usage() :
    print sys.argv[0] + " <options>"
    print "version:1.0"
    print "option:"
    print "\t--handler-template=<handler template file name>"
    print "\t--protocol-template=<protocol template file name>"
    print "\t--init-template=<init template file path>"
    print "\t--enum-template=<enum template file path>"
    print "\t--handler-output=<handler output dir path>"
    print "\t--protocol-output=<protocol output dir path>"
    print "\t--protocol-config=<protocol config file path>"
    print "\t--init-output=<init output dir>"
    print "\t--enum-output=<enum output dir>"
    print "\t--help print usage"

def checkAndFormatDirName(name):
    lastChar = name[len(name) - 1]
    if lastChar != "/" and lastChar != "\\" :
        name += "/"

    return name

def getExt(name):
    fileNameFeilds = name.split('.')
    return fileNameFeilds[len(fileNameFeilds) - 1]

def makeFile(id, name, readFilePath, writeFilePath, fixStr):
        if not os.path.exists(writeFilePath) :
            os.makedirs(writeFilePath)
    
        writeFilePath = checkAndFormatDirName(writeFilePath)

        handlerFile = open(readFilePath)

        fileNameFeilds = readFilePath.split('.')
        ext = fileNameFeilds[len(fileNameFeilds) - 1]

        outFileName = writeFilePath + name.lower() + fixStr  +"." + ext
        if os.path.exists(outFileName):
            return

        handlerFileInfo = handlerFile.read()
        handlerFile.close()

        handlerFileInfo = handlerFileInfo.replace("{@name}", name)
        handlerFileInfo = handlerFileInfo.replace("{@id}", id)

        fp = open(outFileName, "w")
        fp.write(handlerFileInfo)
        fp.close()

def makeFiles():
    global handler_tmp_file
    global protocol_tmp_file
    global init_tmp_file
    global handler_output
    global protocol_output
    global protocol_config
    global init_output
    global enum_output
    global enum_tmp_file

    if not os.path.exists(init_output):
        os.makedirs(init_output)

    init_output = checkAndFormatDirName(init_output)
    enum_output = checkAndFormatDirName(enum_output)


    getLoopZoneRegex = re.compile("\{#begin\}([\s\S]*)\{#end\}", re.M)


    initTmpFile = open(init_tmp_file)
    initInfo = initTmpFile.read()
    initTmpFile.close()

    enumTempFile = open(enum_tmp_file)
    enumInfo = enumTempFile.read()
    enumTempFile.close()

    initResult = ""
    enumResult = ""

    line = 0
    config = open(protocol_config)
    for data in config.readlines():
        if line >= 2:
            feilds=data.split(',')
            if len(feilds) != 2 :
                print "protocol config file invalid"

            id=feilds[0].strip()
            name=feilds[1].strip()

            makeFile(id, name, handler_tmp_file, handler_output, "Handler")
            makeFile(id, name, protocol_tmp_file, protocol_output, "Protoc")

            result = getLoopZoneRegex.findall(initInfo)
            if len(result) != 0:
                info = result[0]
                info = info.replace("{@name}", name)
                info = info.replace("{@id}", id)

                initResult = initResult + info

            result = getLoopZoneRegex.findall(enumInfo)
            if len(result) != 0:
                info = result[0]
                info = info.replace("{@name}", name)
                info = info.replace("{@id}", id)

                enumResult = enumResult + info


        line = line + 1
        
    config.close()

    writeData = getLoopZoneRegex.sub(enumResult, enumInfo)
    ext = getExt(enum_tmp_file)
    enumWriteFile = open(enum_output + "protocEnum." + ext, "w")
    enumWriteFile.write(writeData)
    enumWriteFile.close()
        
    ext = getExt(init_tmp_file)
    outFileName = "protocInit." + ext
    writeData = getLoopZoneRegex.sub(initResult, initInfo)
    initWriteFile = open(init_output + outFileName, "w")
    initWriteFile.write(writeData)
    initWriteFile.close()

    enumInfo = enumInfo.replace("{@name}", name)
    enumInfo = enumInfo.replace("{@id}", id)


if __name__ == "__main__" :
    for key, value in opts:
        if key == "--handler-template" :
            handler_tmp_file = value
        elif key == "--protocol-template":
            protocol_tmp_file = value
        elif key == "--init-template":
            init_tmp_file = value
        elif key == "--handler-output":
            handler_output = value
        elif key == "--protocol-output":
            protocol_output = value
        elif key == "--protocol-config":
            protocol_config = value
        elif key == "--init-output":
            init_output = value
        elif key == "--enum-output":
            enum_output = value
        elif key == "--enum-template":
            enum_tmp_file = value
        elif key == "--help":
            usage()
            sys.exit()

    if handler_tmp_file == None or protocol_tmp_file == None or init_tmp_file == None or handler_output == None or protocol_output == None or protocol_config == None or init_output == None or enum_output == None or enum_tmp_file == None:
        usage()
        sys.exit()

    makeFiles()
