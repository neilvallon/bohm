WaveFunctionCollapse-master:
	wget https://github.com/mxgmn/WaveFunctionCollapse/archive/master.zip
	unzip master.zip WaveFunctionCollapse-master/samples/* WaveFunctionCollapse-master/samples.xml
	rm master.zip

bitmaps: WaveFunctionCollapse-master
	cd WaveFunctionCollapse-master && \
	go run ../cmd/main/main.go -o ../out

clean:
	rm -rf WaveFunctionCollapse-master
	rm -rf out
