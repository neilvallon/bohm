WaveFunctionCollapse-master:
	wget https://github.com/mxgmn/WaveFunctionCollapse/archive/master.zip
	unzip master.zip WaveFunctionCollapse-master/samples/* WaveFunctionCollapse-master/samples.xml
	rm master.zip

bitmaps: WaveFunctionCollapse-master
	go run cmd/main/main.go \
	-i WaveFunctionCollapse-master/samples.xml \
	-t WaveFunctionCollapse-master/samples

clean:
	rm -rf WaveFunctionCollapse-master
	rm -rf out
