AOC

How to build

{go} is your GOPATH, default /home/<user>/go

mkdir -p {go}/src/github.com/skatteetaten
cd {go}/src/github.com/skatteetaten
git clone https://github.com/Skatteetaten/aoc.git
cd aoc
go get
go build

