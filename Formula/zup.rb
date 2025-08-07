class Zup < Formula
  desc "A fast, customizable CLI built with Cobra to automate local development environment setup — from cloning repos to installing tools and running services."
  homepage "https://github.com/devglyph1/homebrew-zup"
  url "https://github.com/devglyph1/homebrew-zup/releases/download/v0.1.6/zup.tar.gz"
  sha256 "c76cabb6377577eb8124f209c3c65ef3215f216613c1d24c849990505962224e"
  version "0.1.6"

  def install
    bin.install "zup"
  end
end
