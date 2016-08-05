require 'mixlib/cli'

module Dapp
  class CLI
    # CLI metadata subcommand
    class Metadata < CLI
      SUBCOMMANDS = %w(flush).freeze

      banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Available subcommands: (for details, dapp SUB-COMMAND --help)

dapp metadata flush
BANNER
    end
  end
end
