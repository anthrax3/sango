#!/bin/sh

for file in images/*/; do
   if [ -d $file ]; then
      if [ -f "$file/config.yml" ]; then
          image=`cat $file/config.yml | grep id: | sed -r "s/id://g" | tr -d ' '`
          echo "$image"
          (cd $file && docker build -t "sango/$image" .);
          docker images -q --filter "dangling=true" | xargs docker rmi
      fi
   fi
done
