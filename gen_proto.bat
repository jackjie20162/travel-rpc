@echo on
set PATH=C:\Users\zzhjz\go\bin;%PATH%
cd /d D:\go_work\simple-admin\travel\travel-rpc
echo Generating proto files...
C:\Users\zzhjz\go\bin\protoc.exe --proto_path=desc --go_out=travel --go_opt=paths=source_relative --go-grpc_out=travel --go-grpc_opt=paths=source_relative desc/travel.proto 2>&1
echo Exit code: %ERRORLEVEL%
dir travel\
