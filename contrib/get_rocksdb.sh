set -e

PWD=$(pwd)
version="6.20.3"
rocksdb="rocksdb"
archive="v${version}.tar.gz"

rm -rf ${rocksdb} ${rocksdb}-${archive}
wget -O ${rocksdb}-${archive} https://github.com/facebook/rocksdb/archive/${archive}
tar -zxvf ${rocksdb}-${archive}
mv ${rocksdb}-${version} ${rocksdb}
