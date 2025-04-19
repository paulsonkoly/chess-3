#!/bin/bash

# credits to https://auscunningham.medium.com/enforcing-git-commit-message-style-b86a45380b0f

commit_message_check(){
      # Get the current branch and apply it to a variable
      currentbranch=`git branch | grep \* | cut -d ' ' -f2`

      # Gets the commits for the current branch and outputs to file
      git log $currentbranch --pretty=format:"%H" --not main > shafile.txt

      failed=0

      # loops through the file an gets the message
      for i in `cat ./shafile.txt`; do 
        # gets the git commit message based on the sha
        git log --format=%B -n 1 "$i" > msgfile.txt

        ####################### TEST STRINGS comment out line 13 to use #########################################
        #gitmessage="feat sdasdsadsaas (AEROGEAR-asdsada)"
        #gitmessage="feat(some txt): some txt (AEROGEAR-****)"
        #gitmessage="docs(some txt): some txt (AEROGEAR-1234)"
        #gitmessage="fix(some txt): some txt (AEROGEAR-5678)"
        #########################################################################################################
        
        # Checks gitmessage for string feat, fix, docs and breaking, if the messagecheck var is empty if fails
        # echo $gitmessage
        # messagecheck=`echo -en $gitmessage | egrep "^bench [0-9]+$"`
        # if [ -z "$messagecheck" ]; then 
        if ! egrep '^bench [0-9]+$' msgfile.txt > /dev/null; then
            echo "Your commit message must contain the bench number"
            failed=1
        fi
      done

      rm shafile.txt  >/dev/null 2>&1

      exit $failed
}

# Calling the function
commit_message_check 
