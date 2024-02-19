// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

uint8 constant SUNDAY = 0;
uint8 constant MONDAY = 1;
uint8 constant TUESDAY = 2;
uint8 constant WEDNESDAY = 3;
uint8 constant THURSDAY = 4;
uint8 constant FRIDAY = 5;
uint8 constant SATURDAY = 6;

uint8 constant SECOND = 0;
uint8 constant MINUTE = 1;
uint8 constant HOUR = 2;
uint8 constant DAY = 3;
uint8 constant WEEK = 4;
uint8 constant MONTH = 5;
uint8 constant YEAR = 6;

struct Party {
    string name;
    address walletAddress;
    bool aware;
}

struct Parties {
    Party application;
    Party proccess;
}

struct Interval {
    uint32 start;
    uint32 end;
}

struct Timeout {
    uint32 increase;
    uint32 end;
}

struct MaxNumberOfOperation {
    uint32 max;
    uint32 used;
    uint32 start;
    uint32 end;
    uint8 timeUnit;
}

struct RightRequestPayment0 {
    MaxNumberOfOperation maxNumberOfOperation0;
    string messageContent0;
    string messageContent1;
}

struct ObligationResponsePayment1 {
    Timeout timeout0;
}

contract PaymentService {
    Parties private parties;
    bool private isActivated = false;

    RightRequestPayment0 private rightRequestPayment0;

    ObligationResponsePayment1 private obligationResponsePayment1;

    mapping(uint8 => uint32) private timeInSeconds;

    event SuccessEvent(string _logMessage);

    modifier onlyProcess() {
        require(isActivated, "This contract is deactivated");
        require(
            parties.proccess.walletAddress == msg.sender,
            "Only the process can execute this operation"
        );
        _;
    }

    modifier onlyParties() {
        require(
            parties.application.walletAddress == msg.sender ||
                parties.proccess.walletAddress == msg.sender,
            "Only the process or the application can execute this operation"
        );
        _;
    }

    constructor(
        Parties memory _parties,
        RightRequestPayment0 memory _rightRequestPayment0,
        ObligationResponsePayment1 memory _obligationResponsePayment1
    ) {
        parties = _parties;

        parties.application.aware = false;
        parties.proccess.aware = false;

        rightRequestPayment0 = _rightRequestPayment0;
        obligationResponsePayment1 = _obligationResponsePayment1;

        rightRequestPayment0.maxNumberOfOperation0.used = 0;
        rightRequestPayment0.maxNumberOfOperation0.start = 0;
        rightRequestPayment0.maxNumberOfOperation0.end = 0;

        timeInSeconds[SECOND] = 1;
        timeInSeconds[MINUTE] = 1 * 60;
        timeInSeconds[HOUR] = 1 * 60 * 60;
        timeInSeconds[DAY] = 1 * 60 * 60 * 24;
        timeInSeconds[WEEK] = 1 * 60 * 60 * 24 * 7;
        timeInSeconds[MONTH] = 1 * 60 * 60 * 24 * 7 * 30;
    }

    function clauseRightRequestPayment0(
        string memory _messageContent0,
        string memory _messageContent1,
        uint32 _accessDateTime
    ) public onlyProcess returns (bool) {
        bool isValid = true;

        bool maxNumberOfOperationIsInitialized0 = rightRequestPayment0
            .maxNumberOfOperation0
            .start ==
            0 &&
            rightRequestPayment0.maxNumberOfOperation0.end == 0;

        bool endPeriodIsLassThanAccessDateTime0 = rightRequestPayment0
            .maxNumberOfOperation0
            .end < _accessDateTime;

        if (
            !maxNumberOfOperationIsInitialized0 ||
            endPeriodIsLassThanAccessDateTime0
        ) {
            rightRequestPayment0.maxNumberOfOperation0.start = _accessDateTime;
            rightRequestPayment0.maxNumberOfOperation0.end =
                _accessDateTime +
                timeInSeconds[
                    rightRequestPayment0.maxNumberOfOperation0.timeUnit
                ];
            rightRequestPayment0.maxNumberOfOperation0.used = 0;
        }

        isValid =
            isValid &&
            rightRequestPayment0.maxNumberOfOperation0.used <=
            rightRequestPayment0.maxNumberOfOperation0.max;

        isValid =
            isValid &&
            keccak256(abi.encodePacked(rightRequestPayment0.messageContent0)) <=
            keccak256(abi.encodePacked(_messageContent0));

        isValid =
            isValid &&
            keccak256(abi.encodePacked(rightRequestPayment0.messageContent1)) <=
            keccak256(abi.encodePacked(_messageContent1));

        obligationResponsePayment1.timeout0.end =
            _accessDateTime +
            obligationResponsePayment1.timeout0.increase;

        require(!isValid, "Error executing clause: RightRequestPayment0");

        emit SuccessEvent("Successful execution!");
        return isValid;
    }

    function clauseObligationResponsePayment1(
        uint32 _accessDateTime
    ) public onlyProcess returns (bool) {
        bool isValid = true;

        isValid =
            isValid &&
            _accessDateTime <= obligationResponsePayment1.timeout0.end;

        require(!isValid, "Error executing clause: ObligationResponsePayment1");

        emit SuccessEvent("Successful execution!");
        return isValid;
    }

    function signContract() public onlyParties returns (bool) {
        if (parties.application.walletAddress == msg.sender) {
            require(
                !parties.application.aware,
                "The contract is already signed"
            );
            parties.application.aware = true;
        }

        if (parties.proccess.walletAddress == msg.sender) {
            require(!parties.proccess.aware, "The contract is already signed");
            parties.proccess.aware = true;
        }

        isActivated = parties.application.aware && parties.proccess.aware;

        return true;
    }
}
