class zup < formula
  desc "A fast, customizable CLI built with Cobra to automate local development environment setup — from cloning repos to installing tools and running services."
  homepage "https://github.com/devglyph1/zup"
  url "https://github.com/devglyph1/zup/releases/download/v0.1.0/zup.tar.gz"
  sha256 "1a8f30860bdce8eea74496dffcee85d25fe83b6a6cd8b743a3dc3804e1129515"
  version "0.1.0"

  def install
    bin.install "zup"
  end
end
