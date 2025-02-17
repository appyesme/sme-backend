git branch --merged | egrep -v "(^\*|main|beta)" | xargs git push origin --delete
git branch --merged | egrep -v "(^\*|main|beta)" | xargs git branch -d
