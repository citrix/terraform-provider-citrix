resource "citrix_policy_set_v2" "example_policy_set_v2" {
    name        = "example-policy-set-v2"
    description = "example-policy-set-v2 for printer assignment policy"
    /* Update the delivery_groups as per your environment 
    * delivery_groups = [
    *     "00000000-0000-0000-0000-000000000000"
    * ]
    */
}

resource "citrix_policy" "example_printer_assignment_policy" {
    policy_set_id   = citrix_policy_set_v2.example_policy_set_v2.id
    name            = "example-printer-assignment-policy"
    description     = "Example Policy with Printer Assignment Policy Setting"
    enabled         = true
}

# citrix_policy_setting resource for "PrinterAssignments"
resource "citrix_policy_setting" "test_printer_printerassignments_PolicySetting" {
    name = "PrinterAssignments"
    policy_id = citrix_policy.example_printer_assignment_policy.id
    use_default = false
    value = jsonencode([
        {
            /* DefaultPrinterOption values
             * 1 - Not Set
             * 2 - Do not adjust
             * 3 - Client Main Printer
             * 4 - Generic Universal Printer
             * 5 - PDF Printer
             */
            "DefaultPrinterOption": 4,
            "SpecificDefaultPrinter": "",
            "SessionPrinters": [
                {
                    "Path": "\\\\printerServer.local\\printerSharedName",
                    "Model": "printerSharedName",
                    "Location": "",
                    "Settings": {
                        /* Only one of `FormName` or `PaperSize` can be overridden each time */
                        "OverrideFormName": false,
                        "FormName": "",
                        "OverridePaperSize": true,
                        /* Paper Size Options:
                        * 1   - Letter
                        * 2   - Letter Small
                        * 3   - Tabloid
                        * 4   - Ledger
                        * 5   - Legal
                        * 6   - Statement
                        * 7   - Executive
                        * 8   - A3
                        * 9   - A4
                        * 10  - A4 Small
                        * 11  - A5
                        * 12  - B4 (JIS)
                        * 13  - B5 (JIS)
                        * 14  - Folio
                        * 15  - Quarto
                        * 16  - 10X14
                        * 17  - 11X17
                        * 18  - Note
                        * 19  - Envelope #9
                        * 20  - Envelope #10
                        * 21  - Envelope #11
                        * 22  - Envelope #12
                        * 23  - Envelope #14
                        * 24  - C Size Sheet
                        * 25  - D Size Sheet
                        * 26  - E Size Sheet
                        * 27  - Envelope DL
                        * 28  - Envelope C5
                        * 29  - Envelope C3
                        * 30  - Envelope C4
                        * 31  - Envelope C6
                        * 32  - Envelope C65
                        * 33  - Envelope B4
                        * 34  - Envelope B5
                        * 35  - Envelope B6
                        * 36  - Envelope Italy
                        * 37  - Envelope Monarch
                        * 38  - Envelope Personal
                        * 39  - US Std Fanfold
                        * 40  - German Std Fanfold
                        * 41  - German Legal Fanfold
                        * 42  - B4 (ISO)
                        * 43  - Japanese Postcard
                        * 44  - 9X11
                        * 45  - 10X11
                        * 46  - 15X11
                        * 47  - Envelope Invite
                        * 50  - Letter Extra
                        * 51  - Legal Extra
                        * 52  - Tabloid Extra
                        * 53  - A4 Extra
                        * 54  - Letter Transverse
                        * 55  - A4 Transverse
                        * 56  - Letter Extra Transverse
                        * 57  - A Plus
                        * 58  - B Plus
                        * 59  - Letter Plus
                        * 60  - A4 Plus
                        * 61  - A5 Transverse
                        * 62  - A3 Extra
                        * 63  - A5 Extra
                        * 64  - B5 (ISO) Extra
                        * 65  - A2
                        * 66  - A3 Transverse
                        * 67  - A3 Extra Transverse
                        * 68  - Japanese Double Postcard
                        * 69  - A6
                        * 70  - Japanese Envelope Kaku #2
                        * 71  - Japanese Envelope Kaku #3
                        * 72  - Japanese Envelope Chou #3
                        * 73  - Japanese Envelope Chou #4
                        * 74  - Letter Rotated
                        * 75  - A3 Rotated
                        * 76  - A4 Rotated
                        * 77  - A5 Rotated
                        * 78  - B4 (JIS) Rotated
                        * 79  - B5 (JIS) Rotated
                        * 80  - Japanese Postcard Rotated
                        * 81  - Japanese Double Postcard Rotated
                        * 82  - A6 Rotated
                        * 83  - Japanese Envelope Kaku #2 Rotated
                        * 84  - Japanese Envelope Kaku #3 Rotated
                        * 85  - Japanese Envelope Chou #3 Rotated
                        * 86  - Japanese Envelope Chou #4 Rotated
                        * 87  - B6 (JIS)
                        * 88  - B6 (JIS) Rotated
                        * 89  - 12X11
                        * 90  - Japanese Envelope You #4
                        * 91  - Japanese Envelope You #4 Rotated
                        * 92  - PRC 16K
                        * 93  - PRC 32K
                        * 94  - PRC 32K Big
                        * 95  - PRC Envelope #1
                        * 96  - PRC Envelope #2
                        * 97  - PRC Envelope #3
                        * 98  - PRC Envelope #4
                        * 99  - PRC Envelope #5
                        * 100 - PRC Envelope #6
                        * 101 - PRC Envelope #7
                        * 102 - PRC Envelope #8
                        * 103 - PRC Envelope #9
                        * 104 - PRC Envelope #10
                        * 105 - PRC16K Rotated
                        * 106 - PRC32K Rotated
                        * 107 - PRC32K Big Rotated
                        * 108 - PRC Envelope #1 Rotated
                        * 109 - PRC Envelope #2 Rotated
                        * 110 - PRC Envelope #3 Rotated
                        * 111 - PRC Envelope #4 Rotated
                        * 112 - PRC Envelope #5 Rotated
                        * 113 - PRC Envelope #6 Rotated
                        * 114 - PRC Envelope #7 Rotated
                        * 115 - PRC Envelope #8 Rotated
                        * 116 - PRC Envelope #9 Rotated
                        * 117 - PRC Envelope #10 Rotated
                        */
                        "PaperSize": 9,
                        "OverridePaperLength": false, // Cannot be changed
                        "Width": 0, // Cannot be changed
                        "Height": 0, // Cannot be changed
                        "OverrideCopyCount": true,
                        "Collated": false, // Can only be set to true if OverrideCopyCount is true
                        "CopyCount": 1,
                        "OverrideScale": true,
                        "Scale": 100,
                        "OverrideColor": true,
                        /* Color Options:
                        * 1 - Monochrome
                        * 2 - Color
                        */
                        "Color": 1,
                        "OverridePrintQuality": true,
                        /* Print Quality Options:
                        * 0  - Custom dpi
                        * -1 - 150 dpi
                        * -2 - 300 dpi
                        * -3 - 600 dpi
                        * -4 - 1200 dpi
                        */
                        "PrintQuality":-4,
                        "XResolution": 0,
                        "YResolution": 0,
                        "OverrideOrientation": true,
                        /* Orientation Options:
                        * 1 - Portrait
                        * 2 - Landscape
                        */
                        "Orientation": 1,
                        "OverrideDuplex": true,
                        /* Duplex Options:
                        * 1 - Simplex
                        * 2 - Horizontal
                        * 3 - Vertical
                        */
                        "Duplex": 2,
                        "OverrideTrueTypeOption": true,
                        /* TrueType Options:
                        * 1 - Bitmap
                        * 2 - Download
                        * 3 - Substitute
                        * 4 - Outline
                        */
                        "TrueTypeOption": 1,
                        "Serialized": "" // Cannot be changed
                    },
                },
                {
                    "Path": "\\\\printerServer.local\\secondPrinter",
                    "Model": "secondPrinter",
                    "Location": "",
                    "Settings": {
                        /* Only one of `FormName` or `PaperSize` can be overridden each time */
                        "OverrideFormName": true,
                        "FormName": "test form",
                        "OverridePaperSize": false,
                        "PaperSize": 0,
                        "OverridePaperLength": false, // Cannot be changed
                        "Width": 0, // Cannot be changed
                        "Height": 0, // Cannot be changed
                        "OverrideCopyCount": true,
                        "Collated": true, 
                        "CopyCount": 1,
                        "OverrideScale": true,
                        "Scale": 50,
                        "OverrideColor": true,
                        "Color": 2,
                        "OverridePrintQuality": true,
                        "PrintQuality":0,
                        "XResolution": 300,
                        "YResolution": 300,
                        "OverrideOrientation": true,
                        "Orientation": 1,
                        "OverrideDuplex": true,
                        "Duplex": 2,
                        "OverrideTrueTypeOption": true,
                        "TrueTypeOption": 1,
                        "Serialized": "" // Cannot be changed
                    },
                }
            ],
            "Filters": ["10.0.0.1", "10.0.0.2", "10.0.0.3"]
        },
        {
            /* DefaultPrinterOption values
             * 1 - Not Set
             * 2 - Do not adjust
             * 3 - Client Main Printer
             * 4 - Generic Universal Printer
             * 5 - PDF Printer
             */
            "DefaultPrinterOption": 5,
            "SpecificDefaultPrinter": "",
            "SessionPrinters": [
                {
                    "Path": "\\\\secondPrinterServer.local\\secondServerFirstPrinter",
                    "Model": "secondServerFirstPrinter",
                    "Location": "",
                    "Settings": {
                        /* Only one of `FormName` or `PaperSize` can be overridden each time */
                        "OverrideFormName": false,
                        "FormName": "",
                        "OverridePaperSize": true,
                        "PaperSize": 9,
                        "OverridePaperLength": false, // Cannot be changed
                        "Width": 0, // Cannot be changed
                        "Height": 0, // Cannot be changed
                        "OverrideCopyCount": true,
                        "Collated": false, // Can only be set to true if OverrideCopyCount is true
                        "CopyCount": 1,
                        "OverrideScale": true,
                        "Scale": 100,
                        "OverrideColor": true,
                        "Color": 1,
                        "OverridePrintQuality": true,
                        "PrintQuality":-4,
                        "XResolution": 0,
                        "YResolution": 0,
                        "OverrideOrientation": true,
                        "Orientation": 1,
                        "OverrideDuplex": true,
                        "Duplex": 2,
                        "OverrideTrueTypeOption": true,
                        "TrueTypeOption": 1,
                        "Serialized": "" // Cannot be changed
                    },
                },
                {
                    "Path": "\\\\secondPrinterServer.local\\secondServerSecondPrinter",
                    "Model": "secondServerSecondPrinter",
                    "Location": "",
                    "Settings": {
                        /* Only one of `FormName` or `PaperSize` can be overridden each time */
                        "OverrideFormName": true,
                        "FormName": "test form",
                        "OverridePaperSize": false,
                        "PaperSize": 0,
                        "OverridePaperLength": false, // Cannot be changed
                        "Width": 0, // Cannot be changed
                        "Height": 0, // Cannot be changed
                        "OverrideCopyCount": true,
                        "Collated": true, 
                        "CopyCount": 1,
                        "OverrideScale": true,
                        "Scale": 50,
                        "OverrideColor": true,
                        "Color": 2,
                        "OverridePrintQuality": true,
                        "PrintQuality":0,
                        "XResolution": 300,
                        "YResolution": 300,
                        "OverrideOrientation": true,
                        "Orientation": 1,
                        "OverrideDuplex": true,
                        "Duplex": 2,
                        "OverrideTrueTypeOption": true,
                        "TrueTypeOption": 1,
                        "Serialized": "" // Cannot be changed
                    },
                }
            ],
            "Filters": ["11.0.0.1", "11.0.0.2", "11.0.0.3"]
        }
    ])
}

