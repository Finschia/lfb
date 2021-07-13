set -e

PWD=$(pwd)
version="1.1.8"
snappy="snappy"
archive="${version}.tar.gz"

rm -rf ${snappy} ${snappy}-${archive}
wget -O ${snappy}-${archive} https://github.com/google/snappy/archive/${archive}
tar -zxvf ${snappy}-${archive}
mv ${snappy}-${version} ${snappy}
