set -e

function check_cleveldb_opts() {
  local ret="false"
  local IFS=" ,"
  for option in $1; do
    if [ "$option" == "cleveldb" ]; then
      ret="true"
      break
    fi
  done
  echo "$ret"
}

install=$(check_cleveldb_opts "$*")
if [ "$install" != "true" ]; then
  echo "Not found 'cleveldb' in args"
  exit 0
fi

PWD=$(pwd)
version="1.23"
leveldb="leveldb"
archive="${version}.tar.gz"

rm -rf ${leveldb} ${leveldb}-${archive}
wget -O ${leveldb}-${archive} https://github.com/google/leveldb/archive/${archive}
tar -zxvf ${leveldb}-${archive}
mv ${leveldb}-${version} ${leveldb}
cd ${leveldb}
mkdir build && cd build
cmake -DCMAKE_BUILD_TYPE=Release -DBUILD_SHARED_LIBS=ON -DLEVELDB_BUILD_TESTS=OFF -DLEVELDB_BUILD_BENCHMARKS=OFF ..
make install
