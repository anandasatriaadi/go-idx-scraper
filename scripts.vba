Sub OpenFilePendukung()
' Turn off screen updating
Application.ScreenUpdating = False

Dim filePath As String
Dim wb As Workbook
Dim sourceSheet As Worksheet
Dim destinationSheet As Worksheet

'===== copy ClosingTrade
    filePath = Range("ClosingTrade.FilePath").Value ' Get the file path and name from the named range
    Set wb = Workbooks.Open(filePath) ' Open the workbook
    Application.WindowState = xlMinimized ' Minimize the opened workbook

    Set sourceSheet = wb.Sheets("InfoTrading")
    Set destinationSheet = ThisWorkbook.Sheets("Closing Trades")

    sourceSheet.Range("A1:G50000").Copy Destination:=destinationSheet.Range("A1")

        Application.CutCopyMode = False

    wb.Close SaveChanges:=False ' Close the workbook without saving changes



'===== copy InfoKurs
    filePath = Range("Quarterly.InfoKurs").Value ' Get the file path and name from the named range
    Set wb = Workbooks.Open(filePath) ' Open the workbook
    Application.WindowState = xlMinimized ' Minimize the opened workbook

    Set sourceSheet = wb.Sheets("InfoKurs")
    Set destinationSheet = ThisWorkbook.Sheets("Info Kurs")

    sourceSheet.Range("B3:D100").Copy Destination:=destinationSheet.Range("A1")

        Application.CutCopyMode = False

    wb.Close SaveChanges:=False ' Close the workbook without saving changes


'===== copy last closing trade
    filePath = Range("LastClosingTrade").Value ' Get the file path and name from the named range
    Set wb = Workbooks.Open(filePath) ' Open the workbook
    Application.WindowState = xlMinimized ' Minimize the opened workbook

    Set sourceSheet = wb.Sheets("Stacked Data")
    Set destinationSheet = ThisWorkbook.Sheets("Last Closing Trade")

    sourceSheet.Range("A:H").Copy Destination:=destinationSheet.Range("A1")
        Application.CutCopyMode = False

    wb.Close SaveChanges:=False ' Close the workbook without saving changes

'===== copy Dividen
    filePath = Range("Dividen.File").Value ' Get the file path and name from the named range
    Set wb = Workbooks.Open(filePath) ' Open the workbook
    Application.WindowState = xlMinimized ' Minimize the opened workbook

    Set sourceSheet = wb.Sheets("Sheet1")
    Set destinationSheet = ThisWorkbook.Sheets("Divident DB")

    sourceSheet.Range("A:G").Copy
    destinationSheet.Range("A1").PasteSpecial xlPasteValues

        Application.CutCopyMode = False

    wb.Close SaveChanges:=False ' Close the workbook without saving changes


    ' Turn on screen updating
    Application.ScreenUpdating = True

End Sub


Sub OpenFileClosingTrade()
' Turn off screen updating
Application.ScreenUpdating = False

Dim filePath As String
Dim wb As Workbook
Dim sourceSheet As Worksheet
Dim destinationSheet As Worksheet


'===== copy last closing trade
    filePath = Range("LastClosingTrade").Value ' Get the file path and name from the named range
    Set wb = Workbooks.Open(filePath) ' Open the workbook
    Application.WindowState = xlMinimized ' Minimize the opened workbook

    Set sourceSheet = wb.Sheets("Stacked Data")
    Set destinationSheet = ThisWorkbook.Sheets("Last Closing Trade")

    sourceSheet.Range("A:H").Copy Destination:=destinationSheet.Range("A1")
        Application.CutCopyMode = False

    wb.Close SaveChanges:=False ' Close the workbook without saving changes


End Sub

Sub OpenFileFromNamedRange()

    'Cancel Filter in ReportDB Sheet
    Sheets("ReportDB").Select
    Rows("4:4").Select

    'Check if data is filtered
    If ActiveSheet.AutoFilterMode And ActiveSheet.FilterMode Then
        ActiveSheet.ShowAllData
    End If

    Range("A4").Select
    Selection.End(xlDown).Select
    ActiveCell.Offset(1, 0).Select

    Sheets("Keys").Select
    Range("A4").Select



    ' Turn off screen updating
    Application.ScreenUpdating = False

    Dim filePath As String
    Dim wb As Workbooks
    Dim sourceSheet As Worksheet
    Dim destinationSheet As Worksheet

    'ensure the file loop while there is a report exist
    Do While Range("ExistReport").Value <> "EMPTY"

    filePath = Range("file.to.open").Value ' Get the file path and name from the named range "lokasi"
    Set wb = Workbooks.Open(filePath) ' Open the workbook
    Application.WindowState = xlMinimized ' Minimize the opened workbook


'===== copy info umum
    Set sourceSheet = wb.Sheets("InfoUmum")    ' Get the "InfoUmum" sheet from the opened workbook
    Set destinationSheet = ThisWorkbook.Sheets("GI")    ' Get the "GeneralInfo" sheet from the current workbook
    destinationSheet.Range("A1:B1000").UnMerge    ' Unmerge the cells in the destination range
    sourceSheet.Range("A1:B1000").Copy Destination:=destinationSheet.Range("A1")    ' Copy the range A1:A100 from the "InfoUmum" sheet to the "GeneralInfo" sheet
    destinationSheet.Range("A1:B1000").UnMerge    ' Unmerge the cells in the destination range



'===== copy Cash Flow
    Set sourceSheet = wb.Sheets("CashFlow")
    Set destinationSheet = ThisWorkbook.Sheets("CF")
    destinationSheet.Range("A1:B1000").UnMerge
    sourceSheet.Range("A1:B1000").Copy Destination:=destinationSheet.Range("A1")
    destinationSheet.Range("A1:B1000").UnMerge

'===== copy Neraca
    Set sourceSheet = wb.Sheets("Neraca")

    Set destinationSheet = ThisWorkbook.Sheets("Banking")
    destinationSheet.Range("A1:B1000").UnMerge
    sourceSheet.Range("A1:B1000").Copy Destination:=destinationSheet.Range("A1")
    destinationSheet.Range("A1:B1000").UnMerge

    Set destinationSheet = ThisWorkbook.Sheets("General")
    destinationSheet.Range("A1:B1000").UnMerge
    sourceSheet.Range("A1:B1000").Copy Destination:=destinationSheet.Range("A1")
    destinationSheet.Range("A1:B1000").UnMerge

    Set destinationSheet = ThisWorkbook.Sheets("Insurance")
    destinationSheet.Range("A1:B1000").UnMerge
    sourceSheet.Range("A1:B1000").Copy Destination:=destinationSheet.Range("A1")
    destinationSheet.Range("A1:B1000").UnMerge

    Set destinationSheet = ThisWorkbook.Sheets("Infrastructure")
    destinationSheet.Range("A1:B1000").UnMerge
    sourceSheet.Range("A1:B1000").Copy Destination:=destinationSheet.Range("A1")
    destinationSheet.Range("A1:B1000").UnMerge

    Set destinationSheet = ThisWorkbook.Sheets("Property")
    destinationSheet.Range("A1:B1000").UnMerge
    sourceSheet.Range("A1:B1000").Copy Destination:=destinationSheet.Range("A1")
    destinationSheet.Range("A1:B1000").UnMerge

    Set destinationSheet = ThisWorkbook.Sheets("Securities")
    destinationSheet.Range("A1:B1000").UnMerge
    sourceSheet.Range("A1:B1000").Copy Destination:=destinationSheet.Range("A1")
    destinationSheet.Range("A1:B1000").UnMerge

    Set destinationSheet = ThisWorkbook.Sheets("Financing")
    destinationSheet.Range("A1:B1000").UnMerge
    sourceSheet.Range("A1:B1000").Copy Destination:=destinationSheet.Range("A1")
    destinationSheet.Range("A1:B1000").UnMerge


'===== copy Rugi Laba
    Set sourceSheet = wb.Sheets("RugiLaba")


    Set destinationSheet = ThisWorkbook.Sheets("Banking")
    destinationSheet.Range("H1:I1000").UnMerge
    sourceSheet.Range("A1:B1000").Copy Destination:=destinationSheet.Range("H1")
    destinationSheet.Range("H1:I1000").UnMerge

    Set destinationSheet = ThisWorkbook.Sheets("General")
    destinationSheet.Range("H1:I1000").UnMerge
    sourceSheet.Range("A1:B1000").Copy Destination:=destinationSheet.Range("H1")
    destinationSheet.Range("H1:I1000").UnMerge

    Set destinationSheet = ThisWorkbook.Sheets("Insurance")
    destinationSheet.Range("H1:I1000").UnMerge
    sourceSheet.Range("A1:B1000").Copy Destination:=destinationSheet.Range("H1")
    destinationSheet.Range("H1:I1000").UnMerge

    Set destinationSheet = ThisWorkbook.Sheets("Infrastructure")
    destinationSheet.Range("H1:I1000").UnMerge
    sourceSheet.Range("A1:B1000").Copy Destination:=destinationSheet.Range("H1")
    destinationSheet.Range("H1:I1000").UnMerge

    Set destinationSheet = ThisWorkbook.Sheets("Property")
    destinationSheet.Range("H1:I1000").UnMerge
    sourceSheet.Range("A1:B1000").Copy Destination:=destinationSheet.Range("H1")
    destinationSheet.Range("H1:I1000").UnMerge

    Set destinationSheet = ThisWorkbook.Sheets("Securities")
    destinationSheet.Range("H1:I1000").UnMerge
    sourceSheet.Range("A1:B1000").Copy Destination:=destinationSheet.Range("H1")
    destinationSheet.Range("H1:I1000").UnMerge

    Set destinationSheet = ThisWorkbook.Sheets("Financing")
    destinationSheet.Range("H1:I1000").UnMerge
    sourceSheet.Range("A1:B1000").Copy Destination:=destinationSheet.Range("H1")
    destinationSheet.Range("H1:I1000").UnMerge




    wb.Close SaveChanges:=False ' Close the workbook without saving changes


'-----------------------
    Dim SummarySheet As Worksheet
    Dim ReportDBSheet As Worksheet
    Dim LastRow As Long

    ' Set the worksheet variables
    Set SummarySheet = ThisWorkbook.Sheets("Summary")
    Set ReportDBSheet = ThisWorkbook.Sheets("ReportDB")

    ' Get the last row in column A of the ReportDB sheet
    LastRow = ReportDBSheet.Cells(Rows.Count, 1).End(xlUp).Row + 1

    ' Copy the data from the EmitenRecord range to the ReportDB sheet
    SummarySheet.Range("EmitenRecord").Copy
    ReportDBSheet.Range("A" & LastRow).PasteSpecial xlPasteValues

    ' Clear the clipboard
    Application.CutCopyMode = False

Loop

    Sheets("Keys").Visible = True

    Sheets("Keys").Select
    Range("B7").Select



' Turn on screen updating
Application.ScreenUpdating = True

' Define visible sheet
        ThisWorkbook.Sheets("Summary").Visible = xlSheetVeryHidden
        ThisWorkbook.Sheets("Infrastructure").Visible = xlSheetVeryHidden
        ThisWorkbook.Sheets("Property").Visible = xlSheetVeryHidden
        ThisWorkbook.Sheets("Last Closing Trade").Visible = xlSheetVeryHidden
        ThisWorkbook.Sheets("Securities").Visible = xlSheetVeryHidden
        ThisWorkbook.Sheets("ReportDB").Visible = xlSheetVisible
        ThisWorkbook.Sheets("Financing").Visible = xlSheetVeryHidden
        ThisWorkbook.Sheets("Calculation Sheet").Visible = xlSheetHidden
        ThisWorkbook.Sheets("Shortlist").Visible = xlSheetVisible
        ThisWorkbook.Sheets("Weed List").Visible = xlSheetHidden
        ThisWorkbook.Sheets("FR Comparison").Visible = xlSheetVisible
        ThisWorkbook.Sheets("Keys").Visible = xlSheetVisible
        ThisWorkbook.Sheets("Divident DB").Visible = xlSheetVeryHidden
        ThisWorkbook.Sheets("Longlist").Visible = xlSheetVisible
        ThisWorkbook.Sheets("Active Portfolio").Visible = xlSheetVisible
        ThisWorkbook.Sheets("LK Q1").Visible = xlSheetVeryHidden
        ThisWorkbook.Sheets("LK Q4").Visible = xlSheetVeryHidden
        'ThisWorkbook.Sheets("Watch This").Visible = xlSheetVeryHidden
        ThisWorkbook.Sheets("Running Profit").Visible = xlSheetHidden
        ThisWorkbook.Sheets("GI").Visible = xlSheetVeryHidden
        ThisWorkbook.Sheets("Banking").Visible = xlSheetVeryHidden
        ThisWorkbook.Sheets("General").Visible = xlSheetVeryHidden
        ThisWorkbook.Sheets("CF").Visible = xlSheetVeryHidden
        ThisWorkbook.Sheets("Closing Trades").Visible = xlSheetVeryHidden
        ThisWorkbook.Sheets("Insurance").Visible = xlSheetVeryHidden
        ThisWorkbook.Sheets("Info Kurs").Visible = xlSheetVeryHidden


    Sheets("FR Comparison").Select
    Range("C2:C3").Select
    Sheets("FR Comparison").Range("C2").Value = "BJBR"


End Sub
