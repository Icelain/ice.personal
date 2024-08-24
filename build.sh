cd src
GOOS=linux GOARCH=amd64 go build 
strip iceblog
mv iceblog ..
echo "built the binary"
cd ..
npx tailwindcss -i ./styles/index.css -o ./static/output.css 
echo "built css"
echo "done"
