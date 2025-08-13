class Zup < Formula
  desc "A fast, customizable CLI built with Cobra to automate local development environment setup — from cloning repos to installing tools and running services."
  homepage "https://github.com/devglyph1/homebrew-zup"
  url "https://github.com/devglyph1/homebrew-zup/releases/download/v0.1.10/zup.tar.gz"
  sha256 "cdae16d05d7acb83e340dffb103b7228e4e9730981376634b3338be60df5bfcf"
  version "0.1.10"

  def install
    bin.install "zup"
  end
end
