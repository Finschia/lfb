set -e

function check_rocksdb_opts() {
  local ret="false"
  local IFS=" ,"
  for option in $1; do
    if [ "$option" == "rocksdb" ]; then
      ret="true"
      break
    fi
  done
  echo "$ret"
}

install=$(check_rocksdb_opts "$*")
if [ "$install" != "true" ]; then
  echo "Not found 'rocksdb' in args"
  exit 0
fi

PWD=$(pwd)
version="6.17.3"
rocksdb="rocksdb"
archive="v${version}.tar.gz"

rm -rf ${rocksdb} ${rocksdb}-${archive}
wget -O ${rocksdb}-${archive} https://github.com/facebook/rocksdb/archive/${archive}
tar -zxvf ${rocksdb}-${archive}
mv ${rocksdb}-${version} ${rocksdb}
cd ${rocksdb}
mkdir build && cd build
cmake -DCMAKE_BUILD_TYPE=Release -DWITH_GFLAGS=0 -DWITH_SNAPPY=1 -DFORCE_AVX2=ON ..
make -j 4 rocksdb-shared

# install, The reason why don't use `make install` is that it builds `static-lib` as well
cp -P ./lib*.so* /usr/local/lib64
cp -r ../include/rocksdb /usr/local/include
