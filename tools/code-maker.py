import re
import sys
import getopt

opts, args = getopt.getopt(sys.argv[1:], "", ["handler-template=", "protocol-template=", "init-template=", "handler-output=", "protocol-output=", "protocol-config=", "init-output=", "ext=", "help"])

handler_tmp_file = None
protocol_tmp_file = None
init_tmp_file = None
handler_output = None
protocol_output = None
protocol_config = None
init_output = None
ext = ""

def usage() :
    print sys.argv[0] + " <options>"
    print "version:1.0"
    print "option:"
    print "\t--handler-template=<handler template file name>"
    print "\t--protocol-template=<protocol template file name>"
    print "\t--init-template=<init template file path>"
    print "\t--handler-output=<handler output dir path>"
    print "\t--protocol-output=<protocol output dir path>"
    print "\t--protocol-config=<protocol config file path>"
    print "\t--init-output=<init output dir>"
    print "\t--help print usage"

def makeFile(id, name, readFilePath, writeFilePath):
        handlerFile = open(readFilePath)

        fileNameFeilds = readFilePath.split('.')
        ext = fileNameFeilds[len(fileNameFeilds) - 1]

        handlerFileInfo = handlerFile.read()
        handlerFile.close()

        handlerFileInfo = handlerFileInfo.replace("{@name}", name)
        handlerFileInfo = handlerFileInfo.replace("{@id}", id)
        outFileName = name.lower() + "Handler." + ext
        fp = open(outFileName, "w")
        fp.write(handlerFileInfo)
        fp.close()
    

def makeFiles():
    getLoopZoneRegex = re.compile("\{#begin\}([\s\S]*)\{#end\}", re.M)

    initTempFile = open(init_tmp_file)
    config = open(protocol_config)

    initInfo = initTempFile.read()
    initResult = ""
    for data in config.readlines():
        feilds=data.split(',')
        if len(feilds) != 2 :
            print "protocol config file invalid"

        id=feilds[0].strip()
        name=feilds[1].strip()

        makeFile(id, name, handler_tmp_file, handler_output)
        makeFile(id, name, protocol_tmp_file, protocol_output)

        result = getLoopZoneRegex.findall(initInfo)
        if len(result) != 0:
            info = result[0]
            info = info.replace("{@name}", name)
            info = info.replace("{@id}", id)

            initResult = initResult + info
        
        print getLoopZoneRegex.sub(initResult, initInfo)
        

    config.close()
    initTempFile.close()


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
    elif key == "--help":
        usage()
        sys.exit()

makeFiles()

#if handler_tmp_file == None or protocol_tmp_file == None or init_tmp_file == None or handler_output == None or protocol_output == None or protocol_config == None or init_output == None or ext == None :
#    Usage()
#    sys.exit()
