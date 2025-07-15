class Zup < Formula
  desc "A fast, customizable CLI built with Cobra to automate local development environment setup — from cloning repos to installing tools and running services."
  homepage "https://github.com/devglyph1/zup"
  url "https://github.com/devglyph1/zup/releases/download/v0.1.1/zup.tar.gz"
  sha256 "b1124c9c226ba9b24948118f61686e0c925d53a56081ec4e9c36abb55af17728"
  version "0.1.1"

  def install
    bin.install "zup"
  end
end
