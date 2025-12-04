import logging
import asyncio
# NEW IMPORT PATH for v3.x
from pymodbus.server import StartAsyncTcpServer
from pymodbus.datastore import ModbusSequentialDataBlock, ModbusDeviceContext, ModbusServerContext
from pymodbus import ModbusDeviceIdentification

# Configuration
PORT = 5020

async def run_server():
    # --- 1. Define the LOGO! 8 Memory Map ---
    # Inputs (I1-I24) -> Discrete Inputs 0-23
    # d0
    inputs_block = ModbusSequentialDataBlock(1, [0] + [1]*2 + [0])

    # Outputs (Q1-Q20) & Flags (M1-M64) -> Coils
    # Start at 8192 (Q1). Cover up to 8319 (M64).
    coils_block = ModbusSequentialDataBlock(8193, [0]*128)

    # Analog Inputs (AI1-AI8) -> Input Registers 0-7
    # i@0
    analog_inputs_block = ModbusSequentialDataBlock(1032, [500, 850, 350, 0, 0, 0, 0, 200])
    vm_block = ModbusSequentialDataBlock(0, [112] + [0]*424 + [250])

    # --- 2. Create the Slave Context ---
    store = ModbusDeviceContext(
        di=inputs_block,
        co=coils_block,
        hr=vm_block,
        ir=analog_inputs_block,
    )

    # 'single=True' means this context applies to all Unit IDs (Slave IDs)
    context = ModbusServerContext(devices=store, single=True)

    # --- 3. Server Identity ---
    identity = ModbusDeviceIdentification()
    identity.VendorName = 'Simulated Siemens'
    identity.ProductName = 'Mock LOGO! Server'
    identity.ModelName = 'MockServer'
    identity.MajorMinorRevision = '3.0'

    # --- 4. Start the Server (v3.x Style) ---
    print(f"Starting Mock LOGO! 8 Server on 0.0.0.0:{PORT}...")

    # In v3.x, StartAsyncTcpServer is an awaitable coroutine
    await StartAsyncTcpServer(context=context, identity=identity, address=("0.0.0.0", PORT))

if __name__ == "__main__":
    try:
        # v3.x requires explicit asyncio loop handling
        asyncio.run(run_server())
    except PermissionError:
        print(f"Error: Permission denied on port {PORT}. Try running as admin or change to port 5020.")
    except Exception as e:
        print(f"Error: {e}")
