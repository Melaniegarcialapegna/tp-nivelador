
AMOUNT_BYTES_UINT32 = 4
BIRTHDATE_LENGTH = 10 #YYYY-MM-DD
AMOUNT_BYTES_CONST = AMOUNT_BYTES_UINT32 + AMOUNT_BYTES_UINT32 + AMOUNT_BYTES_UINT32 + BIRTHDATE_LENGTH

# Serializes a bet to a byte array
# For the fields that are dinamic in size, will be used a separator to know how many bytes to read for each field
# like long_dinamic_field_i|dinamic_field_i|
def serializeBet():
    return 

def writeDynamicField(): 
    return 

#Deserializes a byte array to a bet
def deserializeBet():
    return


def readDynamicField():
	return
