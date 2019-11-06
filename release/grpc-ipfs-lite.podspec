Pod::Spec.new do |spec|
    spec.name         = "grpc-ipfs-lite"
    spec.version      = "<version>"
    spec.summary      = "gRPC server exposing API for ipfs-lite"
    spec.description  = <<-DESC
                        Objective C framework for grpc-ipfs-lite. You should not usually use this pod directly, but instead use the ipfs-lite pod.
                      DESC
    spec.homepage     = "https://github.com/textileio/grpc-ipfs-lite"
    spec.license      = "MIT"
    spec.author       = { "textile.io" => "contact@textile.io" }
    spec.platform     = :ios, "7.0"
    spec.source       = { :http => 'https://github.com/textileio/grpc-ipfs-lite/releases/download/v<version>/grpc-ipfs-lite_v<version>_ios-framework.tar.gz' }
    spec.vendored_frameworks = 'Mobile.framework'
    spec.requires_arc = false
    spec.pod_target_xcconfig = { 'GCC_PREPROCESSOR_DEFINITIONS' => '$(inherited) GPB_USE_PROTOBUF_FRAMEWORK_IMPORTS=1', 'OTHER_LDFLAGS[arch=i386]' => '-Wl,-read_only_relocs,suppress' }
  end