#!/bin/sh

WORKDIR=$1
BUILDDIR=$2

# Based on gist.github.com/gboudreau/install-ffmpeg-amazon-linux.sh
# and https://trac.ffmpeg.org/wiki/CompilationGuide/Centos
if [ "`/usr/bin/whoami`" != "root" ]; then
    echo "You need to execute this script as root."
    exit 1
fi

cat > /etc/yum.repos.d/centos.repo<<EOF
[centos]
name=CentOS-6 – Base
baseurl=http://mirror.centos.org/centos/6/os/x86_64/
gpgcheck=1
gpgkey=http://mirror.centos.org/centos/RPM-GPG-KEY-CentOS-6
enabled=1
priority=1
protect=1
includepkgs=SDL SDL-devel gsm gsm-devel libtheora theora-tools
EOF
rpm --import http://mirror.centos.org/centos/RPM-GPG-KEY-CentOS-6

#rpm -Uhv http://ec2-23-22-86-129.compute-1.amazonaws.com/pub/sam/1.3/el6/x86_64/SAM_brew_latest/toplink/packages/libraw1394/2.0.4/1.el6/x86_64/libraw1394-2.0.4-1.el6.x86_64.rpm

rpm -Uhv http://pkgs.repoforge.org/rpmforge-release/rpmforge-release-0.5.3-1.el6.rf.x86_64.rpm
yum -y update
yum -y install libraw1394

yum -y install wget
yum -y install glibc gcc gcc-c++ autoconf automake libtool git make nasm pkgconfig
yum -y install SDL-devel a52dec a52dec-devel alsa-lib-devel faac faac-devel faad2 faad2-devel
yum -y install freetype-devel giflib gsm gsm-devel imlib2 imlib2-devel lame lame-devel libICE-devel libSM-devel libX11-devel
yum -y install libXau-devel libXdmcp-devel libXext-devel libXrandr-devel libXrender-devel libXt-devel
yum -y install libogg libvorbis vorbis-tools mesa-libGL-devel mesa-libGLU-devel xorg-x11-proto-devel zlib-devel
yum -y install libtheora theora-tools
yum -y install ncurses-devel
yum -y install libdc1394 libdc1394-devel
yum -y install amrnb-devel amrwb-devel opencore-amr-devel
yum -y install xz

yum -y remove yasm
cd $WORKDIR
rm -rf yasm-1.2.0
wget http://www.tortall.net/projects/yasm/releases/yasm-1.2.0.tar.gz
tar xzfv yasm-1.2.0.tar.gz && rm -f yasm-1.2.0.tar.gz

wget http://www.nasm.us/pub/nasm/releasebuilds/2.13.01/nasm-2.13.01.tar.xz
tar -xf nasm-2.13.01.tar.xz

git clone git://git.videolan.org/x264.git

rm -rf ffmpeg
git clone git://source.ffmpeg.org/ffmpeg.git
cd ffmpeg
#git checkout 2a111c99a60fdf4fe5eea2b073901630190c6c93
git checkout n4.0.2

cd $WORKDIR
cd yasm-1.2.0
./configure --prefix="$BUILDDIR" --bindir="$HOME/bin" && make install
export "PATH=$PATH:$HOME/bin" 


cd $WORKDIR
cd nasm-2.13.01
./configure --prefix=/usr && make && make install

# fetch latest libx264 and install
cd $WORKDIR
cd x264
echo "Building x264..."
./configure --prefix="$BUILDDIR" --enable-pic --enable-static && make && make install
#ffmpeg static linking of libx264

# install nasm
# install libx264 
cd $WORKDIR
cd ffmpeg

PKG_CONFIG_PATH="$BUILDDIR/lib/pkgconfig" ./configure --prefix="$BUILDDIR" \
--extra-cflags="-I$BUILDDIR/include -Bstatic" \
--extra-ldflags="-L$BUILDDIR/lib -ldl -Bstatic" \
--bindir="$HOME/bin" \
--pkg-config-flags="--static" \
--enable-gpl \
--enable-libx264
make
make install

cp $HOME/bin/ffmpeg /usr/bin/
cp $HOME/bin/ffprobe /usr/bin/
cp $HOME/bin/ffmpeg /workdir/
# Test the resulting ffmpeg binary
ffmpeg -version
