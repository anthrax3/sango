#!/bin/sh

curl -O http://localhost:8080/jnlpJars/jenkins-cli.jar

for file in ../images/[!_]*/; do
   if [ -d $file ]; then
      if [ -f "$file/config.yml" ]; then
          image=`cat $file/config.yml | grep id: | sed "s/id://g" | tr -d ' '`
          xml=`cat template.xml | sed "s/{{name}}/$image/g"`
	  if java -jar jenkins-cli.jar -s http://localhost:8080 get-job sango-$image > /dev/null ; then
	      echo "$xml" | java -jar jenkins-cli.jar -s http://localhost:8080 update-job sango-$image 
	  else
	      echo "$xml" | java -jar jenkins-cli.jar -s http://localhost:8080 create-job sango-$image
	  fi
      fi
   fi
done
