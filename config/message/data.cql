CREATE TABLE IF NOT EXISTS graduation_thesis.CONV_MSG (
    conv_id text,
    conv_msg_id bigint,
    msg_time bigint, 
    sender text,
    content blob,
    iv text,
    PRIMARY KEY (conv_id, conv_msg_id)
) WITH CLUSTERING ORDER BY (conv_msg_id DESC)
 AND compaction={
    'compaction_window_size': '30',
    'compaction_window_unit': 'DAYS',
    'class': 'org.apache.cassandra.db.compaction.TimeWindowCompactionStrategy'
 };

CREATE TABLE IF NOT EXISTS graduation_thesis.USER_INBOX (
    user_id text,
    inbox_msg_id bigint, 
    conv_id text,
    conv_msg_id bigint,
    msg_time bigint,
    sender text,
    content blob,
    iv text,
    PRIMARY KEY (user_id, conv_id, conv_msg_id)
) WITH CLUSTERING ORDER BY (conv_id DESC, conv_msg_id DESC)
AND default_time_to_live = 2592000
AND compaction={
    'compaction_window_size': '10',
    'compaction_window_unit': 'DAYS',
    'class': 'org.apache.cassandra.db.compaction.TimeWindowCompactionStrategy'
};

CREATE TABLE IF NOT EXISTS graduation_thesis.READ_RECEIPT (
    conv_id text,
    user_id text,
    last_seen_msg bigint,
    PRIMARY KEY (conv_id, user_id)
); 

CREATE TABLE IF NOT EXISTS graduation_thesis.LASTSEEN (
    user_id text,
    status text
);

CREATE TABLE USER_INBOX {
    user_id text,
    conv_id text,
    last_msg_time bigint,
    PRIMARY KEY (user_id, last_msg_time)
}
