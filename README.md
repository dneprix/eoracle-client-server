# Eoracle Client-Server Application

A scalable client-server application built in Go that implements an ordered map data structure with O(1) operations, using RabbitMQ for message queuing and supporting concurrent command execution.

## Architecture

### Server
- Configurable via command-line flags
- Maintains an in-memory Ordered Map data structure with O(1) add, get, delete operations and thread-safe implementation RWMutex
- Reads commands from two RabbitMQ queues. First queue for read commands and second queue for write commands
- Executes commands concurrently using worker goroutines. Use two types (read/white) of worker pools for parallel command execution
- Outputs results to a file
- Supports read operations (get, getall) running in parallel without blocking
- Handle command execution errors and retry proccessing RabbitMQ messsage
- Graceful Shutdown: stop listening queue, stop workers, finish processing commands, stop server
- Validation for commands and values


### Client
- Configurable via command-line flags
- Supports command line mode
- Sends commands to two RabbitMQ queues. First queue for read commands and second queue for write commands
- Multiple clients can run simultaneously
- Validation for commands and values

### Client Commands
- `add <key> <value>`: Add/update key-value pair
- `delete <key>`: Remove key-value pair
- `get <key>`: Retrieve value for key
- `getall`: Get all key-value pairs in insertion order

## Quick Start

1. **Build binaries**:
   ```bash
   make build   
   ```
   or 

   ```bash
   make build-arm64   
   ```

2. **Run RabbitMQ**:
   ```bash
   rabbitmq-start
   ```

3. **Run the server**:
   ```bash
    ./bin/server
   ```

4. **Run client**:
   ```bash
   ./bin/client
   ```

## Usage Examples

```bash
> add user1 john
> add user2 jane
> get user1
> getall
> delete user1
```

## Configuration

### Server Options
- `-rabbit-url`: RabbitMQ connection URL (default: `amqp://guest:guest@localhost:5672/`)
- `-read-queue-name`: Queue name (default: `read-commands`)
- `-write-queue-name`: Queue name (default: `write-commands`)
- `-output-file`: Output file path (default: `server_output.txt`)
- `-read-workers`: Number of read worker goroutines (default: `10`)
- `-write-workers`: Number of write worker goroutines (default: `10`)


### Client Options
- `-rabbit-url`: RabbitMQ connection URL
- `-read-queue-name`: Queue name (default: `read-commands`)
- `-write-queue-name`: Queue name (default: `write-commands`)


## Testing
Run the unit tests with race flag:
```bash
make test
```
## Monitoring

RabbitMQ Management UI is available at http://localhost:15672 (guest/guest)

Server logs show:
- Command processing status
- Worker activity
- Error conditions
- Performance metrics

## Design Decisions

### Ordered Map Implementation
- Uses a combination of hash map and doubly-linked list
- Hash map provides O(1) key-based access
- Linked list maintains insertion order
- Thread-safe with RWMutex for concurrent access
- Read operations can run in parallel, writes are exclusive

### Concurrency Strategy
- Worker pool pattern for command processing
- Two types of worker pools (read and write operations)
- Unbuffered channels for command distribution
- RWMutex allows multiple concurrent reads
- File output is synchronized with mutex

### Scalability Features
- Configurable size for worker pools 
- Queue-based decoupling of clients and server
- Stateless client design

## Assumptions
1. **Message ordering**: Commands are processed in parallel, so strict ordering is not guaranteed across different operations
2. **Persistence**: Data is stored in memory only; server restart will lose all data
3. **Error handling**: Failed commands are logged but don't stop the server
4. **Network reliability**: RabbitMQ provides message durability and delivery guarantees
5. **Key-value types**: Both keys and values are strings as specified
6. **File output**: Results are appended to output file
7. **Only one server instance**: Running multiple server instances was not required. We can run multiple server instance but with different RabbitMQ queue and different Ordered Map data for each server instance.
0. **Binary versions**: No need support for binary versions (Ex. v0.0.1, v0.0.2, ..  etc)

## Future Enhancements for production ready solution

1. **Persistence**: Add optional disk-based storage
2. **Binary versions**: Add support versions for client/server builds and binaries (Ex. v0.0.1, v0.0.2, ..  etc)
3. **Health checks**: Add health check endpoints
4. **Load balancing**: Add support for multiple server instances
5. **Authentication**: Add authentication for queue access
6. **Compression**: Add message compression for large payloads
7. **Idempotency**: Add support `Idempotency` for RabbitMQ messages, add dedpulicator for RabbitMQ messages, support requestID/traceID
8. **Logging**: Add support log levels [info, warn, error, debug]
9. **Multiple queues**: For some performance cases let's think to use sepate queue for each command type. 
9. **Security**: Validate command values for security injections
