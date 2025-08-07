class Zup < Formula
  desc "A fast, customizable CLI built with Cobra to automate local development environment setup — from cloning repos to installing tools and running services."
  homepage "https://github.com/devglyph1/homebrew-zup"
  url "https://github.com/devglyph1/homebrew-zup/releases/download/v0.1.7/zup.tar.gz"
  sha256 "ed218fa90e8adeecfe3a888f70393ba985d8bd4754bb664f95bebece097e9c73"
  version "0.1.7"

  def install
    bin.install "zup"
  end
end
