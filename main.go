/*
    A simple single user Spreadsheet which allows two operations.
    setCellValue and getCellValue
    1) setCellValue(cellId, Value)
    2) getCellValue(cellId)
    
    cellId is of the format "<Alphabet in caps><Row Number>"
    Note:
    - Alphabet in caps corresponds to the column.
    - Row Number is > 1
    - Value is string represnetation of an integer or a mathematical formula.
    - Formula starts with =
    
    Assumptions:
    - Max number of columns: 26
    - Formula supports only addition and subtraction of cell IDs and numbers. Ex: "=A1+B2-C3+10"
    - Formula supports range sum. Ex: A1:A5, A1:C4 etc
    - Example formula with additon, subtraction and range: "=A1+B2-C3+10+A2:B3"
    - There is no cyclic dependency on the cell. Example: formula of A1 cannot be "=B1" and formula
      of B1 cannot be "=A1" at the same time.
    - By default, the value of each cell is 0.
*/

package main

import (
    "errors"
    "fmt"
    "strings"
    "strconv"
)

type Cell struct {
    // List of cells that are dependent on this cell. If this cell value is updated,
    // values of all the cells that dependent on this cell are updated simultaneously
    // for displaying real time updated values of the affected cells.
    //
    // Note: Map data structure is used instead of a list for O(1) search/deletions.
    dependentCells map[string]interface{}
    
    // Integer value of the cell. This is displayed in the UI.
    value *int
    
    // Formula of the cell.
    formula *string
}

type SpreadSheet struct {
    // Spreadsheet is a matrix of cells.
    cells [][]*Cell
}

type CellId struct {
    row, col int
    sign string
    val *int
}

func CreateSpreadSheet(numRows, numCols int) *SpreadSheet {
    sheet := new(SpreadSheet)
    sheet.cells = make([][]*Cell, numRows)
    if numCols > 26 {
        // Set max cols to 26.
        numCols = 26
    }

    for i := 0; i < numRows; i++ {
        sheet.cells[i] = make([]*Cell, numCols)
        for j := 0; j < numCols; j++ {
            sheet.cells[i][j] = new(Cell)
            sheet.cells[i][j].dependentCells = make(map[string]interface{})
            value := 0
            sheet.cells[i][j].value = &value
        }
    }
    
    return sheet
}

func (sheet *SpreadSheet) SetCellValue(cellId string, value string) error {
    row, col, err := getCellRowCol(cellId)
    if err != nil {
        return err
    }
    
    if len(strings.TrimSpace(value)) == 0 {
        value = "0"
    }
    
    // Remove dependees.
    if sheet.cells[row][col].formula != nil {
        sheet.deleteDependees(cellId, *sheet.cells[row][col].formula)
    }

    valueInt, err := strconv.Atoi(value)
    if err == nil {
        sheet.cells[row][col].value = &valueInt
        // If value is an integer, unset the formula.
        sheet.cells[row][col].formula = nil
    } else {
        sheet.cells[row][col].formula = &value
        sheet.computeCellValue(cellId)
    }
    
    // Add dependees.
    if sheet.cells[row][col].formula != nil {
        sheet.addDependees(cellId, *sheet.cells[row][col].formula)
    }
    
    // Recompute dependents value. This is because the cells whose value depends
    // on this cell will have a stale value.
    for cid := range sheet.cells[row][col].dependentCells {
        sheet.computeCellValue(cid)
    }
    return nil
}

// Function that returns the value of the cell.
func (sheet *SpreadSheet) GetCellValue(cellId string) (int, error) {
    row, col, err := getCellRowCol(cellId)
    if err != nil {
        return 0, err
    }
 
    if row >= len(sheet.cells) {
        errMsg := "Row number out of bounds in cellId"
        fmt.Println(errMsg)
        return 0, errors.New(errMsg)
    }
    
    if col >= len(sheet.cells[0]) {
        errMsg := "Column value out of bounds in cellId"
        fmt.Println(errMsg)
        return 0, errors.New(errMsg)
    }

    return *sheet.cells[row][col].value, nil
}

// Returns row, col numbers and nil if cell ID is valid. Else returns -1, -1, and error.
//
// Cell ID is valid if first character (column) is a capital alphabet and rest of the characters (row) are a string
// representation of an integer.
func getCellRowCol(cellId string) (int, int, error) {
    col := int(cellId[0]-'A')
    if col < 0 || col >= 26 {
        errMsg := "Invalid col number in cellId"
        fmt.Println(errMsg)
        return -1, -1, errors.New(errMsg) 
    }
    row, err := strconv.Atoi(cellId[1:])
    if err != nil {
        errMsg := "Invalid row number in cellId"
        fmt.Println(errMsg)
        return -1, -1, errors.New(errMsg)
    }
    
    return row-1, col, nil
}

// Function to get the cell IDs in a given range. 
// For example, if rangeStr is A1:B2, then A1, A2, B1, B2 are returned.
func getCellIdsFromRange(rangeStr, sign string) []*CellId {
    cellIds := make([]*CellId, 0)
    if !strings.Contains(rangeStr, ":") {
        cellId := new(CellId)
        cellId.sign = sign
        
        val, err := strconv.Atoi(rangeStr)
        if err == nil {
            cellId.val = &val
        } else {
            cellId.row, cellId.col, _ = getCellRowCol(rangeStr)
        }
        cellIds = append(cellIds, cellId)
    } else {
        cells := strings.Split(rangeStr, ":")
        topRow, leftCol, _ := getCellRowCol(cells[0])
        bottomRow, rightCol, _ := getCellRowCol(cells[1])
        for r := topRow; r <= bottomRow; r++ {
            for c := leftCol; c <= rightCol; c++ {
                cellId := &CellId{
                    sign: sign,
                    row: r,
                    col: c,
                }
                cellIds = append(cellIds, cellId)
            }
        }
    }
    
    return cellIds
}

// Function to get all cell IDs in a formula.
func getCellIdsFromFormula(formula string) []*CellId {
    cellIds := make([]*CellId, 0)
    
    // Remove the leading =.
    formula = formula[1:]
    start := 0
    sign := "+"
    for i := 0; i < len(formula); i++ {
        if formula[i] != '+' && formula[i] != '-' {
            continue
        }

        cellIds = append(cellIds, getCellIdsFromRange(formula[start:i], sign)...)
        sign = string(formula[i])
        start = i+1
    }
    
    cellIds = append(cellIds, getCellIdsFromRange(formula[start:], sign)...)
    return cellIds
}

// Function to delete cellId from the dependents map of each cell ID in the formula.
func (sheet *SpreadSheet) deleteDependees(cellId, formula string) {
    cellIds := getCellIdsFromFormula(formula)
    for _, id := range cellIds {
        delete(sheet.cells[id.row][id.col].dependentCells, cellId)
    }
}

// Function to add cellId to the dependents map of each cell ID in the formula.
func (sheet *SpreadSheet) addDependees(cellId, formula string) {
    cellIds := getCellIdsFromFormula(formula)
    for _, id := range cellIds {
        sheet.cells[id.row][id.col].dependentCells[cellId] = true
    }
}

// Function takes cell ID and compute the value from the formula.
func (sheet *SpreadSheet) computeCellValue(cellId string) {
    row, col, err := getCellRowCol(cellId)
    if err != nil {
        return
    }
    value := 0
    formula := sheet.cells[row][col].formula
    if formula == nil {
        sheet.cells[row][col].value = &value
        return
    }
    
    cellIds := getCellIdsFromFormula(*formula)
    for _, id := range cellIds {
        if id.sign == "+" {
            if id.val != nil {
                value += *id.val
            } else {
                value += *sheet.cells[id.row][id.col].value
            }
        } else if id.sign == "-" {
            if id.val != nil {
                value -= *id.val
            } else {
                value -= *sheet.cells[id.row][id.col].value
            } 
        }
    }
    
    // Iterate over the formula and compute the val.
    sheet.cells[row][col].value = &value
}


func main() {
    sheet := CreateSpreadSheet(3,3)
    
    // Base case.
    sheet.SetCellValue("A1","10")
    fmt.Println(sheet.GetCellValue("A1")) // 10
    fmt.Println(sheet.GetCellValue("C3")) // 0
    
    // Set C3 to A1+A2+B1+B2+C1+C2. Note A2, B1, B2, C1, C2 are not set.
    sheet.SetCellValue("C3", "=A1:C2") // C3 
    fmt.Println(sheet.GetCellValue("C3")) // 10
    
    // Updating C2 should update the value of C3 because 
    // formula of C3 depends on C2.
    sheet.SetCellValue("C2", "=A1")
    fmt.Println(sheet.GetCellValue("C3")) // 20
    
    sheet.SetCellValue("A2", "5")
    sheet.SetCellValue("B2", "=A1+A2") // 15
    sheet.SetCellValue("C1", "=A1-A2+5") // 10
    
    fmt.Println(sheet.GetCellValue("C1")) // 10
    
    // Updating the A2, B2, C1 should update the value of C3.
    fmt.Println(sheet.GetCellValue("C3")) // 50
    
    // Remove the formula of C3 by setting a static value.
    sheet.SetCellValue("C3", "25")
    fmt.Println(sheet.GetCellValue("C3")) // 25
}
