set -e

PWD=$(pwd)
version="1.23"
leveldb="leveldb"
archive="${version}.tar.gz"

rm -rf ${leveldb} ${leveldb}-${archive}
wget -O ${leveldb}-${archive} https://github.com/google/leveldb/archive/${archive}
tar -zxvf ${leveldb}-${archive}
mv ${leveldb}-${version} ${leveldb}
