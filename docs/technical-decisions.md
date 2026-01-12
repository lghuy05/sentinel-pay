1. Why Microservices?

    Each fraud detection method can scale independently

    Can deploy ML separately from rules engine

    Different teams can own different services

2. Why Kafka?

    Real-time event streaming for fraud detection

    Decouples services (no direct API calls)

    Can replay events for debugging

    Handles high throughput (MoMo scale)

3. Why Kubernetes?

    Auto-scaling during Tet holiday peaks

    Multi-region deployment (Hanoi/HCMC)

    Rolling updates without downtime

    Service mesh for traffic management

4. Why Python for ML?

    ML ecosystem better in Python

    Can swap models easily (TensorFlow, PyTorch, etc.)

    Separate from Java business logic

    Can use GPU for training if needed


Technical Highlights:

    Event-driven architecture with Kafka

    Microservices with circuit breakers

    Real-time feature engineering

    ML model serving in production

    Vietnamese market understanding
