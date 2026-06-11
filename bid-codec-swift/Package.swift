// swift-tools-version: 5.9

import PackageDescription

let package = Package(
    name: "BidCodec",
    products: [
        .library(name: "BidCodec", targets: ["BidCodec"]),
        .executable(name: "BidCodecVectorRunner", targets: ["BidCodecVectorRunner"]),
    ],
    targets: [
        .target(name: "BidCodec"),
        .executableTarget(name: "BidCodecVectorRunner", dependencies: ["BidCodec"]),
    ]
)
