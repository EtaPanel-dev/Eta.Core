// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract LogStorage {
    struct LogEntry {
        uint256 id;
        address sender;
        uint256 timestamp;
        string level;      // INFO, WARN, ERROR, DEBUG
        string source;     // 日志来源/模块
        string message;    // 日志内容
        bytes32 hash;      // 内容哈希，用于验证
        bool exists;
    }

    // 存储所有日志条目
    mapping(uint256 => LogEntry) public logs;

    // 按发送者索引日志
    mapping(address => uint256[]) public logsByAddress;

    // 按日志级别索引
    mapping(string => uint256[]) public logsByLevel;

    // 按来源索引
    mapping(string => uint256[]) public logsBySource;

    // 按时间范围索引 (天级别)
    mapping(uint256 => uint256[]) public logsByDay;

    uint256 public totalLogs;
    uint256 public constant MAX_MESSAGE_LENGTH = 1000;

    event LogStored(
        uint256 indexed logId,
        address indexed sender,
        string indexed level,
        string source,
        uint256 timestamp
    );

    event BatchLogStored(
        uint256 indexed startId,
        uint256 indexed endId,
        address indexed sender,
        uint256 count
    );

    modifier validLogLevel(string memory _level) {
        require(
            keccak256(bytes(_level)) == keccak256(bytes("INFO")) ||
            keccak256(bytes(_level)) == keccak256(bytes("WARN")) ||
            keccak256(bytes(_level)) == keccak256(bytes("ERROR")) ||
            keccak256(bytes(_level)) == keccak256(bytes("DEBUG")),
            "Invalid log level"
        );
        _;
    }

    modifier validMessage(string memory _message) {
        require(bytes(_message).length <= MAX_MESSAGE_LENGTH, "Message too long");
        require(bytes(_message).length > 0, "Message cannot be empty");
        _;
    }

    // 存储单条日志
    function storeLog(
        string memory _level,
        string memory _source,
        string memory _message
    )
        public
        validLogLevel(_level)
        validMessage(_message)
        returns (uint256)
    {
        uint256 logId = totalLogs++;
        bytes32 contentHash = keccak256(abi.encodePacked(_level, _source, _message, block.timestamp));

        logs[logId] = LogEntry({
            id: logId,
            sender: msg.sender,
            timestamp: block.timestamp,
            level: _level,
            source: _source,
            message: _message,
            hash: contentHash,
            exists: true
        });

        // 更新索引
        logsByAddress[msg.sender].push(logId);
        logsByLevel[_level].push(logId);
        logsBySource[_source].push(logId);

        // 按天索引 (timestamp / 86400 = 天数)
        uint256 day = block.timestamp / 86400;
        logsByDay[day].push(logId);

        emit LogStored(logId, msg.sender, _level, _source, block.timestamp);

        return logId;
    }

    // 批量存储日志 (节省gas)
    function storeBatchLogs(
        string[] memory _levels,
        string[] memory _sources,
        string[] memory _messages
    )
        public
        returns (uint256[] memory)
    {
        require(_levels.length == _sources.length && _sources.length == _messages.length,
                "Arrays length mismatch");
        require(_levels.length > 0 && _levels.length <= 50, "Invalid batch size");

        uint256[] memory logIds = new uint256[](_levels.length);
        uint256 startId = totalLogs;

        for (uint256 i = 0; i < _levels.length; i++) {
            require(bytes(_messages[i]).length <= MAX_MESSAGE_LENGTH && bytes(_messages[i]).length > 0,
                    "Invalid message");

            uint256 logId = totalLogs++;
            bytes32 contentHash = keccak256(abi.encodePacked(_levels[i], _sources[i], _messages[i], block.timestamp));

            logs[logId] = LogEntry({
                id: logId,
                sender: msg.sender,
                timestamp: block.timestamp,
                level: _levels[i],
                source: _sources[i],
                message: _messages[i],
                hash: contentHash,
                exists: true
            });

            // 更新索引
            logsByAddress[msg.sender].push(logId);
            logsByLevel[_levels[i]].push(logId);
            logsBySource[_sources[i]].push(logId);

            uint256 day = block.timestamp / 86400;
            logsByDay[day].push(logId);

            logIds[i] = logId;
        }

        emit BatchLogStored(startId, totalLogs - 1, msg.sender, _levels.length);

        return logIds;
    }

    // 获取单条日志
    function getLog(uint256 _logId)
        public
        view
        returns (LogEntry memory)
    {
        require(logs[_logId].exists, "Log does not exist");
        return logs[_logId];
    }

    // 获取用户的日志ID列表
    function getLogsByAddress(address _address)
        public
        view
        returns (uint256[] memory)
    {
        return logsByAddress[_address];
    }

    // 获取指定级别的日志ID列表
    function getLogsByLevel(string memory _level)
        public
        view
        returns (uint256[] memory)
    {
        return logsByLevel[_level];
    }

    // 获取指定来源的日志ID列表
    function getLogsBySource(string memory _source)
        public
        view
        returns (uint256[] memory)
    {
        return logsBySource[_source];
    }

    // 获取指定日期的日志ID列表
    function getLogsByDay(uint256 _day)
        public
        view
        returns (uint256[] memory)
    {
        return logsByDay[_day];
    }

    // 获取日志总数
    function getTotalLogs() public view returns (uint256) {
        return totalLogs;
    }

    // 分页获取日志
    function getLogsPaginated(uint256 _offset, uint256 _limit)
        public
        view
        returns (LogEntry[] memory)
    {
        require(_limit > 0 && _limit <= 100, "Invalid limit");
        require(_offset < totalLogs, "Offset out of range");

        uint256 end = _offset + _limit;
        if (end > totalLogs) {
            end = totalLogs;
        }

        LogEntry[] memory result = new LogEntry[](end - _offset);

        for (uint256 i = _offset; i < end; i++) {
            if (logs[i].exists) {
                result[i - _offset] = logs[i];
            }
        }

        return result;
    }

    // 验证日志内容完整性
    function verifyLog(uint256 _logId)
        public
        view
        returns (bool)
    {
        require(logs[_logId].exists, "Log does not exist");

        LogEntry memory log = logs[_logId];
        bytes32 expectedHash = keccak256(abi.encodePacked(log.level, log.source, log.message, log.timestamp));

        return log.hash == expectedHash;
    }
}