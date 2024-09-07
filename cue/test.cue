// Serviceの定義
#Service: {
    apiVersion: "v1"
    kind:       "Service"
    metadata: {
        name: string
        namespace?: string
        labels?: [string]: string
    }
    spec: {
        type: "ClusterIP" | "NodePort" | "LoadBalancer" | "ExternalName"
        selector?: [string]: string
        ports: [...#Port]
    }
}

// ポートの定義
#Port: {
    name?: string
    protocol: "TCP" | "UDP" | *"TCP"
    port: int
    targetPort?: int
    nodePort?: int
}

// 実際のServiceインスタンス
exampleService: #Service & {
    metadata: {
        name: "my-service"
        labels: {
            "app": "my-app"
        }
    }
    spec: {
        type: "ClusterIP"
        selector: {
            "app": "my-app"
        }
        ports: [{
            name: "http"
            port: 80
        }, {
            name: "https"
            port: 443
        }]
    }
}
