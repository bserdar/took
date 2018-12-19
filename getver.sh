#!?bin/sh

tag=`git tag -l --points-at HEAD`
if [ -z $tag ] ; then
    tag=`git rev-parse --verify --short HEAD`
    tag="dev.$tag"
fi
echo $tag

