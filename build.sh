cd src
GOOS=linux GOARCH=amd64 go build
GOOS=linux GOARCH=arm64 go build -o iceblogarm
strip iceblog
mv iceblog ..
mv iceblogarm ..
echo "built the binaries"
cd ..
npx tailwindcss -i ./styles/index.css -o ./static/output.css 
echo "built css"
echo "done"
