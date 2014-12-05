#!/bin/sh

tgz=`mktemp -d`/sango.tar.gz
tar czf $tgz --exclude-vcs .

rm -rf .build
mkdir .build
cp -r images .build
cd .build

(cd images/_base && docker build -t sango/_base .);

for file in images/[!_]*/; do
   if [ -d $file ]; then
      if [ -f "$file/config.yml" ]; then
          image=`cat $file/config.yml | grep id: | sed -r "s/id://g" | tr -d ' '`
          echo "$image"
          (cd $file && cp $tgz . && docker build -t sango/$image . && rm sango.tar.gz);
          docker images -q --filter "dangling=true" | xargs docker rmi
      fi
   fi
done

rm -rf .build
