#!/bin/sh

tgz=`mktemp -d`/sango.tar.gz
tar czf $tgz --exclude-vcs .

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
