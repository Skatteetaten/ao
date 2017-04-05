```
           ,ggg,     _,gggggg,_          ,gggg,  
          dP""8I   ,d8P""d8P"Y8b,      ,88"""Y8b,
         dP   88  ,d8'   Y8   "8b,dP  d8"     `Y8
        dP    88  d8'    `Ybaaad88P' d8'   8b  d8
       ,8'    88  8P       `""""Y8  ,8I    "Y88P'
       d88888888  8b            d8  I8'          
 __   ,8"     88  Y8,          ,8P  d8           
dP"  ,8P      Y8  `Y8,        ,8P'  Y8,          
Yb,_,dP       `8b, `Y8b,,__,,d8P'   `Yba,,_____, 
 "Y8P"         `Y8   `"Y8888P"'       `"Y8888888
 
 ```
 
 
 #Build?
 
 ```
{go} is your GOPATH, default /home/<user>/go

mkdir -p {go}/src/github.com/skatteetaten
cd {go}/src/github.com/skatteetaten
git clone https://github.com/Skatteetaten/aoc.git
cd aoc
go get
go build
 
 ```
 
 
 #Dependencies?
 
 ```
 glide up # Update dependencies. Only run when you change something in glide.yaml
 glide install # install dependencies
 ```