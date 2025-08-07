class Zup < Formula
  desc "A fast, customizable CLI built with Cobra to automate local development environment setup — from cloning repos to installing tools and running services."
  homepage "https://github.com/devglyph1/homebrew-zup"
  url "https://github.com/devglyph1/homebrew-zup/releases/download/v0.1.8/zup.tar.gz"
  sha256 "25690e92162ff7787609622daa35c9b5d9df1b58a19474882d1d71f4984ffc53"
  version "0.1.8"

  def install
    bin.install "zup"
  end
end
