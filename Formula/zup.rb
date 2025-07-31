class Zup < Formula
  desc "A fast, customizable CLI built with Cobra to automate local development environment setup — from cloning repos to installing tools and running services."
  homepage "https://github.com/devglyph1/homebrew-zup"
  url "https://github.com/devglyph1/homebrew-zup/releases/download/v0.1.2/zup.tar.gz"
  sha256 "75fca130ecea01bab59c596f2f23d6e9966ee01dd700d2ed07d4869388ddfd8a"
  version "0.1.3"

  def install
    bin.install "zup"
  end
end
