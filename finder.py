#!/usr/bin/python
import os
import sys
# import yaml
# print 'Argument List:', str(sys.argv[0])
# print len(sys.argv)-1

target_key_words=['---']

input_number=len(sys.argv)
if input_number>0:
    for i in range(1,input_number):
        target_key_words[i-1]=str(sys.argv[i])

print target_key_words

rootdir = './'
for subdir, dirs, files in os.walk(rootdir):
    for file in files:
        target_file=os.path.join(subdir, file)
        if (target_key_words[0] in open(target_file).read()):
           print('>>>>> >>>>>' + target_file)
