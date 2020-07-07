import os
import sys
import re
import json
import nbformat
from nbconvert.preprocessors import (
  ExecutePreprocessor,
  CellExecutionError
)
from papermill import (
    execute as mill,
    PapermillException
)
from jupyter_client.kernelspec import NoSuchKernel

def readNotebook(nb_file_path):
  """
  Opens a Jupyter Notbook to read
   - Parameters: nb_file_path
   - Returns: notebook, boolean
  """
  f = None
  nb = None
  success = False
  try:
    f = open(nb_file_path, "r")
    nb = nbformat.read(f, as_version=4)
    success = True
  except (UnicodeDecodeError, FileNotFoundError) as err:
    print(err, file=sys.stderr)
  finally:
    if f != None:
      f.close()
  return nb, success

def writeNotebook(nb_file_path, nb):
  """
  Opens a Jupyter Notbook to write
   - Parameters: nb_file_path, nb
   - Returns: boolean
  """
  f = None
  success = False
  try:
    f = open(nb_file_path, "w")
    nbformat.write(nb, f)
    success = True
  except UnicodeDecodeError as err:
    print(err, file=sys.stderr)
  finally:
    if f != None:
      f.close()
    return success

def parameterize(nb_input_file_path, nb_output_file_name, parameters, nb):
  """
  Adds parameters to the Jupyter Notebook
   - Parameters: nb_input_file_path, nb_output_file_name, parameters, nb
   - Returns: notebook, boolean
  """
  success = False
  try:
    nb = mill.execute_notebook(
      nb_input_file_path,
      nb_output_file_name,
      parameters=parameters,
      prepare_only=True,
      log_output=True
    )
    success = True
  except FileNotFoundError as err:
    print(err, file=sys.stderr)
  finally:
    return nb, success

def startPreprocessor(kernel):
  """
  Opens a Jupyter Notbook to write
   - Parameters: kernel
   - Returns: preprocessor, boolean
  """
  ex = None
  success = False
  try:
    ex = ExecutePreprocessor(kernel_name=kernel)
    ex.km, ex.kc = ex.start_new_kernel()
    success = True
  except NoSuchKernel as err:
    print(err, file=sys.stderr)
    ex = None
  finally:
    return ex, success

if __name__ == "__main__":
  # Return Codes
  ERROR_ARGUMENT     = 1
  ERROR_INPUT        = 2
  ERROR_PAPERMILL    = 3
  ERROR_PREPROCESSOR = 4
  ERROR_GENERIC_CELL = 5
  ERROR_OUTPUT       = 6

  # Arguments
  if len(sys.argv) != 5:
    print("ERROR> Incorrect Usage")
    print("Incorrect number of arguments", file=sys.stderr)
    sys.exit(ERROR_ARGUMENT)
  nb_input_file_path = sys.argv[1]
  input_file_path = sys.argv[2]
  rotate_degrees = sys.argv[3]
  output_directory = sys.argv[4]
  
  kernel = "python3"
  nb_output_file_path = os.path.join(output_directory, f'EXECUTED_{os.path.basename(nb_input_file_path)}')

  # Open Notebook
  nb, is_read = readNotebook(nb_input_file_path)
  if is_read == False:
    print("ERROR> Input File not Found or isn't Valid!")
    sys.exit(ERROR_INPUT)
  print("LOG> Read in Notebook!")

  # Parameters
  parameters = dict(source_image_path=input_file_path, rotate_degrees=rotate_degrees, output_dir=output_directory)
  nb, parameters_set = parameterize(nb_input_file_path, nb_output_file_path, parameters, nb)
  if parameters_set == False:
    print("ERROR> Papermill Failed to Set to Parameters!")
    sys.exit(ERROR_PAPERMILL)
  print("LOG> Parameters Set!")

  # Setup Preprocessor
  ex, is_start = startPreprocessor(kernel)
  if is_start == False:
    print("ERROR> ExecutePreprocessor Failed to Start!")
    sys.exit(ERROR_PREPROCESSOR)
  print("LOG> Kernel Started!")

  # Execute Notebook
  return_code = 0
  cell_count = 0
  try:
    for cell in nb.cells:
      ex.preprocess_cell(cell, None, cell_count)  
      cell_count += 1
    print("LOG> Execution Complete!")
  except CellExecutionError as e: 
    # Clean Exception Info
    ipyReturn = str(e)
    ansi_escape = re.compile(r'(\x9B|\x1B\[)[0-?]*[ -\/]*[@-~]')
    cleanedReturn = ansi_escape.sub('', ipyReturn)

    # Find Error Code
    metadata = cell.metadata
    if "code" in metadata:
      return_code = metadata["code"]
    else:
      return_code = ERROR_GENERIC_CELL

    # Output
    print(cleanedReturn, file=sys.stderr)
    print("ERROR> Execution Failed at Cell " + str(cell_count) + "!")
  finally:
    # Write Output to Notebook
    if writeNotebook(nb_output_file_path, nb) == False:
      print("ERROR> Output File is Invalid or isn't Writeable!")
      sys.exit(ERROR_OUTPUT)
    print("LOG> Writing Complete!")

    # Shutdown Kernel and Exit
    ex.km.shutdown_kernel() 
    print("LOG> Quitting...")
    sys.exit(return_code)