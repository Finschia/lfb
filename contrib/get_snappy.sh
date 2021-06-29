set -e

function check_cleveldb_or_rocksdb_opts() {
  local ret="false"
  local IFS=" ,"
  for option in $1; do
    if [ "$option" == "cleveldb" ] || [ "$option" == "rocksdb" ]; then
      ret="true"
      break
    fi
  done
  echo "$ret"
}

install=$(check_cleveldb_or_rocksdb_opts "$*")
if [ "$install" != "true" ]; then
  echo "Not found 'cleveldb' nor 'rocksdb' in args"
  exit 0
fi

PWD=$(pwd)
version="1.1.8"
snappy="snappy"
archive="${version}.tar.gz"

rm -rf ${snappy} ${snappy}-${archive}
wget -O ${snappy}-${archive} https://github.com/google/snappy/archive/${archive}
tar -zxvf ${snappy}-${archive}
mv ${snappy}-${version} ${snappy}
cd ${snappy}
mkdir build && cd build
cmake -DBUILD_SHARED_LIBS=ON -DSNAPPY_BUILD_TESTS=OFF -DSNAPPY_REQUIRE_AVX2=ON ..
make install
