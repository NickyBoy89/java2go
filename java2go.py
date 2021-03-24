import coloredlogs, logging
import argparse

import javaparser
import gowriter

# Create a logger object.
log = logging.getLogger(__name__)

coloredlogs.install(level='DEBUG')

parser = argparse.ArgumentParser(description='Converts java files to Golang files, given an input and output directory')
parser.add_argument('-in', dest='indir', type=str, help='Directory to find the input files')
parser.add_argument('-out', dest='outdir', type=str, help='Directory to output the directory files')

args = parser.parse_args()

def convert(input):

    goClasses = ""

    for javaClasses in javaparser.parseClasses(input):
        goClasses += gowriter.writeClassToFile(javaClasses, javaparser.parseMethodsAndVariables(input, javaClasses["name"]))

    return goClasses
