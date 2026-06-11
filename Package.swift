// swift-tools-version: 5.9

import PackageDescription

let package = Package(
    name: "bid754",
    products: [
        .library(name: "BidCodec", targets: ["BidCodec"]),
        .executable(name: "BidCodecVectorRunner", targets: ["BidCodecVectorRunner"]),
    ],
    targets: [
        .target(name: "BidCodec", path: "bid754-codec-swift/Sources/BidCodec"),
        .executableTarget(
            name: "BidCodecVectorRunner",
            dependencies: ["BidCodec"],
            path: "bid754-codec-swift/Sources/BidCodecVectorRunner"
        ),
    ]
)
