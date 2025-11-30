(Taken from Microchip's datasheet.)

# PIC16(L)F1773/6 INSTRUCTION SET

## BYTE-ORIENTED FILE REGISTER OPERATIONS

| Mnemonic, Operands | Description | Cycles | 14-Bit Opcode MSb | 14-Bit Opcode LSb | Status Affected |
|-------------------|-------------|--------|-------------------|-------------------|-----------------|
| ADDWF f, d | Add W and f | 1 | 00 0111 dfff | ffff | C, DC, Z |
| ADDWFC f, d | Add with Carry W and f | 1 | 11 1101 dfff | ffff | C, DC, Z |
| ANDWF f, d | AND W with f | 1 | 00 0101 dfff | ffff | Z |
| ASRF f, d | Arithmetic Right Shift | 1 | 11 0111 dfff | ffff | C, Z |
| LSLF f, d | Logical Left Shift | 1 | 11 0101 dfff | ffff | C, Z |
| LSRF f, d | Logical Right Shift | 1 | 11 0110 dfff | ffff | C, Z |
| CLRF f | Clear f | 1 | 00 0001 1fff | ffff | Z |
| CLRW – | Clear W | 1 | 00 0001 0000 | 00xx | Z |
| COMF f, d | Complement f | 1 | 00 1001 dfff | ffff | Z |
| DECF f, d | Decrement f | 1 | 00 0011 dfff | ffff | Z |
| INCF f, d | Increment f | 1 | 00 1010 dfff | ffff | Z |
| IORWF f, d | Inclusive OR W with f | 1 | 00 0100 dfff | ffff | Z |
| MOVF f, d | Move f | 1 | 00 1000 dfff | ffff | Z |
| MOVWF f | Move W to f | 1 | 00 0000 1fff | ffff | |
| RLF f, d | Rotate Left f through Carry | 1 | 00 1101 dfff | ffff | C |
| RRF f, d | Rotate Right f through Carry | 1 | 00 1100 dfff | ffff | C |
| SUBWF f, d | Subtract W from f | 1 | 00 0010 dfff | ffff | C, DC, Z |
| SUBWFB f, d | Subtract with Borrow W from f | 1 | 11 1011 dfff | ffff | C, DC, Z |
| SWAPF f, d | Swap nibbles in f | 1 | 00 1110 dfff | ffff | |
| XORWF f, d | Exclusive OR W with f | 1 | 00 0110 dfff | ffff | Z |

## BYTE ORIENTED SKIP OPERATIONS

| Mnemonic, Operands | Description | Cycles | 14-Bit Opcode MSb | 14-Bit Opcode LSb | Status Affected |
|-------------------|-------------|--------|-------------------|-------------------|-----------------|
| DECFSZ f, d | Decrement f, Skip if 0 | 1(2) | 00 1011 dfff | ffff | |
| INCFSZ f, d | Increment f, Skip if 0 | 1(2) | 00 1111 dfff | ffff | |

## BIT-ORIENTED FILE REGISTER OPERATIONS

| Mnemonic, Operands | Description | Cycles | 14-Bit Opcode MSb | 14-Bit Opcode LSb | Status Affected |
|-------------------|-------------|--------|-------------------|-------------------|-----------------|
| BCF f, b | Bit Clear f | 1 | 01 00bb bfff | ffff | |
| BSF f, b | Bit Set f | 1 | 01 01bb bfff | ffff | |

## BIT-ORIENTED SKIP OPERATIONS

| Mnemonic, Operands | Description | Cycles | 14-Bit Opcode MSb | 14-Bit Opcode LSb | Status Affected |
|-------------------|-------------|--------|-------------------|-------------------|-----------------|
| BTFSC f, b | Bit Test f, Skip if Clear | 1 (2) | 01 10bb bfff | ffff | |
| BTFSS f, b | Bit Test f, Skip if Set | 1 (2) | 01 11bb bfff | ffff | |

## LITERAL OPERATIONS

| Mnemonic, Operands | Description | Cycles | 14-Bit Opcode MSb | 14-Bit Opcode LSb | Status Affected |
|-------------------|-------------|--------|-------------------|-------------------|-----------------|
| ADDLW k | Add literal and W | 1 | 11 1110 kkkk | kkkk | C, DC, Z |
| ANDLW k | AND literal with W | 1 | 11 1001 kkkk | kkkk | Z |
| IORLW k | Inclusive OR literal with W | 1 | 11 1000 kkkk | kkkk | Z |
| MOVLB k | Move literal to BSR | 1 | 00 0000 001k | kkkk | |
| MOVLP k | Move literal to PCLATH | 1 | 11 0001 1kkk | kkkk | |
| MOVLW k | Move literal to W | 1 | 11 0000 kkkk | kkkk | |
| SUBLW k | Subtract W from literal | 1 | 11 1100 kkkk | kkkk | C, DC, Z |
| XORLW k | Exclusive OR literal with W | 1 | 11 1010 kkkk | kkkk | Z |

## CONTROL OPERATIONS

| Mnemonic, Operands | Description | Cycles | 14-Bit Opcode MSb | 14-Bit Opcode LSb | Status Affected |
|-------------------|-------------|--------|-------------------|-------------------|-----------------|
| BRA k | Relative Branch | 2 | 11 001k kkkk | kkkk | |
| BRW – | Relative Branch with W | 2 | 00 0000 0000 | 1011 | |
| CALL k | Call Subroutine | 2 | 10 0kkk kkkk | kkkk | |
| CALLW – | Call Subroutine with W | 2 | 00 0000 0000 | 1010 | |
| GOTO k | Go to address | 2 | 10 1kkk kkkk | kkkk | |
| RETFIE - | Return from interrupt | 2 | 00 0000 0000 | 1001 | |
| RETLW k | Return with literal in W | 2 | 11 0100 kkkk | kkkk | |
| RETURN – | Return from Subroutine | 2 | 00 0000 0000 | 1000 | |

## INHERENT OPERATIONS

| Mnemonic, Operands | Description | Cycles | 14-Bit Opcode MSb | 14-Bit Opcode LSb | Status Affected |
|-------------------|-------------|--------|-------------------|-------------------|-----------------|
| CLRWDT – | Clear Watchdog Timer | 1 | 00 0000 0110 | 0100 | TO, PD |
| NOP – | No Operation | 1 | 00 0000 0000 | 0000 | |
| OPTION – | Load OPTION_REG register with W | 1 | 00 0000 0110 | 0010 | |
| RESET – | Software device Reset | 1 | 00 0000 0000 | 0001 | |
| SLEEP – | Go into Standby mode | 1 | 00 0000 0110 | 0011 | TO, PD |
| TRIS f | Load TRIS register with W | 1 | 00 0000 0110 | 0fff | |

## C-COMPILER OPTIMIZED

| Mnemonic, Operands | Description | Cycles | 14-Bit Opcode MSb | 14-Bit Opcode LSb | Status Affected |
|-------------------|-------------|--------|-------------------|-------------------|-----------------|
| ADDFSR n, k | Add Literal k to FSRn | 1 | 11 0001 0nkk | kkkk | |
| MOVIW n mm | Move Indirect FSRn to W with pre/post inc/dec modifier, mm | 1 | 00 0000 0001 | 0nmm | Z |
| MOVIW k[n] | Move INDFn to W, Indexed Indirect. | 1 | 11 1111 0nkk | kkkk | Z |
| MOVWI n mm | Move W to Indirect FSRn with pre/post inc/dec modifier, mm | 1 | 00 0000 0001 | 1nmm | |
| MOVWI k[n] | Move W to INDFn, Indexed Indirect. | 1 | 11 1111 1nkk | kkkk | |