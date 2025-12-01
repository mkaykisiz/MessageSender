// MongoDB Initialization Script
// This script creates the necessary database, collections, and indexes

// Switch to sender database
use sender_db;

// Create messages collection if it doesn't exist
db.createCollection("messages");

// Create indexes for efficient querying

// Index for efficient querying of unsent messages (pending/failed)
// This index is used by the worker to fetch messages that need to be sent
// The compound index on status and created_at allows efficient filtering and sorting
db.messages.createIndex(
    { "status": 1, "created_at": 1 },
    {
        name: "idx_status_created_at",
        background: true
    }
);

// Index for retrieving sent messages
// This index is used by the retrieve-sent-messages API endpoint
db.messages.createIndex(
    { "status": 1 },
    {
        name: "idx_status",
        background: true
    }
);

// Optional: Index for recipient lookup (if needed for future features)
db.messages.createIndex(
    { "recipient": 1 },
    {
        name: "idx_recipient",
        background: true,
        sparse: true
    }
);

// Print created indexes
print("Created indexes:");
db.messages.getIndexes().forEach(function (index) {
    printjson(index);
});

print("\nDatabase initialization completed successfully!");
