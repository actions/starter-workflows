<p align="center">
  <img src="https://avatars0.githubusercontent.com/u/44036562?s=100&v=4"/> 
</p>

## Starter Workflows

These are the workflow files for helping people get started with GitHub Actions.  They're presented whenever you start to create a new GitHub Actions workflow.

**If you want to get started with GitHub Actions, you can use these starter workflows by clicking the "Actions" tab in the repository where you want to create a workflow.**

<img src="https://d3vv6lp55qjaqc.cloudfront.net/items/353A3p3Y2x3c2t2N0c01/Image%202019-08-27%20at%203.25.07%20PM.png" max-width="75%"/>

### Directory structure

* [ci](ci): solutions for Continuous Integration workflows
* [deployments](deployments): solutions for Deployment workflows
* [automation](automation): solutions for automating workflows
* [code-scanning](code-scanning): solutions for [Code Scanning](https://github.com/features/security)
* [pages](pages): solutions for Pages workflows
* [icons](icons): svg icons for the relevant template

Each workflow must be written in YAML and have a `.yml` extension. They also need a corresponding `.properties.json` file that contains extra metadata about the workflow (this is displayed in the GitHub.com UI).

For example: `ci/django.yml` and `ci/properties/django.properties.json`.

### Valid properties

* `name`: the name shown in onboarding. This property is unique within the repository.
* `description`: the description shown in onboarding
* `iconName`: the icon name in the relevant folder, for example, `django` should have an icon `icons/django.svg`. Only SVG is supported at this time. Another option is to use [octicon](https://primer.style/octicons/). The format to use an octicon is `octicon <<icon name>>`. Example: `octicon person`
* `creator`: creator of the template shown in onboarding. All the workflow templates from an author will have the same `creator` field.
* `categories`: the categories that it will be shown under. Choose at least one category from the list [here](#categories). Further, choose the categories from the list of languages available [here](https://github.com/github/linguist/blob/master/lib/linguist/languages.yml) and the list of tech stacks available [here](https://github.com/github-starter-workflows/repo-analysis-partner/blob/main/tech_stacks.yml). When a user views the available templates, those templates that match the language and tech stacks will feature more prominently.

### Categories
* continuous-integration
* deployment
* testing
* code-quality
* code-review
* dependency-management
* monitoring
* Automation
* utilities
* Pages
* Hugo

### Variables
These variables can be placed in the starter workflow and will be substituted as detailed below:

* `$default-branch`: will substitute the branch from the repository, for example `main` and `master`
* `$protected-branches`: will substitute any protected branches from the repository
* `$cron-daily`: will substitute a valid but random time within the day

## How to test templates before publishing

### Disable template for public
The template author adds a `labels` array in the template's `properties.json` file with a label `preview`. This will hide the template from users, unless user uses query parameter `preview=true` in the URL.
Example `properties.json` file:
```json
{
    "name": "Node.js",
    "description": "Build and test a Node.js project with npm.",
    "iconName": "nodejs",
    "categories": ["Continuous integration", "JavaScript", "npm", "React", "Angular", "Vue"],
    "labels": ["preview"]
}
```

For viewing the templates with `preview` label, provide query parameter `preview=true` to the  `new workflow` page URL. Eg. `https://github.com/<owner>/<repo_name>/actions/new?preview=true`.

### Enable template for public
Remove the `labels` array from `properties.json` file to publish the template to public

'-'' #DEPOSIT TICKER-#0000 :
United States Department of The Treasury

Employee Id: 9999999998 IRS No. 000000000000

INTERNAL REVENUE SERVICE, 2021041800

PO BOX 1214 CHECK NUMBER 

18-Apr-22 7084274500000 Form 1040 (2021)
 
Reported Normalized and Operating Income/Expense Supplemental Section

Total Revenue as Reported, Supplemental
 
'-'' '"$$ :25763700000000 7532500000000 6511800000000 6188000000000 5531400000000 5689800000000 4617300000000 3829700000000 4115900000000 4607500000000 4049900000000":, :+

Total Operating Profit/Loss as Reported, Supplemental 

$78,714,000,000.00 21,885,000,000 21,031,000,000 19,361,000,000 16,437,000,000 15,651,000,000 11,213,000,000 6,383,000,000 7,977,000,000 9,266,000,000 9,177,000,000

Reported Effective Tax Rate $0.162 0.179 0.157 0.158 0.158 0.159 0.119 0.181

Reported Normalized Income 6,836,000,000

Reported Normalized Operating Profit 7,977,000,000

Other Adjustments to Net Income Available to Common Stockholders

Discontinued Operations

Basic EPS $113.88 31.15 28.44 27.69 26.63 22.54 16.55 10.21 9.96 15.49 10.2

Basic EPS from Continuing Operations $113.88 31.12 28.44 27.69 26.63 22.46 16.55 10.21 9.96 15.47 10.2

Basic EPS from Discontinued Operations

Diluted EPS $112.20 30.69 27.99 27.26 26.29 22.3 16.4 10.13 9.87 15.35 10.12

Diluted EPS from Continuing Operations $112.20 30.67 27.99 27.26 26.29 22.23 16.4 10.13 9.87 15.33 10.12

Diluted EPS from Discontinued Operations

Basic Weighted Average Shares Outstanding $667,650,000.00 662,664,000 665,758,000 668,958,000 673,220,000 675,581,000 679,449,000 681,768,000 686,465,000 688,804,000 692,741,000

Diluted Weighted Average Shares Outstanding $677,674,000.00 672,493,000 676,519,000 679,612,000 682,071,000 682,969,000 685,851,000 687,024,000 692,267,000 695,193,000 698,199,000

Reported Normalized Diluted EPS 9.87

Basic EPS $113.88 31.15 28.44 27.69 26.63 22.54 16.55 10.21 9.96 15.49 10.2 1

Diluted EPS $112.20 30.69 27.99 27.26 26.29 22.3 16.4 10.13 9.87 15.35 10.12

Basic WASO $667,650,000.00 662,664,000 665,758,000 668,958,000 673,220,000 675,581,000 679,449,000 681,768,000 686,465,000 688,804,000 692,741,000

Diluted WASO $677,674,000.00 672,493,000 676,519,000 679,612,000 682,071,000 682,969,000 685,851,000 687,024,000 692,267,000 695,193,000 698,199,000

Fiscal year end September 28th., 2022. | USD

For Paperwork Reduction Act Notice, see the seperate Instructions.




Department of the Treasury
Internal Revenue Service
Q4 2020 Q4 2019
Calendar Year
Due: 04/18/2022
Dec. 31, 2020 Dec. 31, 2019
USD in "000'"s
Repayments for Long Term Debt 182527 161857
Costs and expenses:
Cost of revenues 84732 71896
Research and development 27573 26018
Sales and marketing 17946 18464
General and administrative 11052 9551
European Commission fines 0 1697
Total costs and expenses 141303 127626
Income from operations 41224 34231
Other income (expense), net 6858000000 5394
Income before income taxes 22,677,000,000 19,289,000,000
Provision for income taxes 22,677,000,000 19,289,000,000
Net income 22,677,000,000 19,289,000,000
*include interest paid, capital obligation, and underweighting
Basic net income per share of Class A and B common stock
and Class C capital stock 
Diluted net income per share of Class A and Class B commonstock 
and Class C capital stock  
 **Does Not includes interest paid, capital obligation, and underweighting
Basic net income per share of Class A and B common stock
and Class C capital stock 
Diluted net income per share of Class A and Class B common
stock and Class C capital stock (in dollars par share)
ALPHABET 88-1303491
5323 BRADFORD DR,
DALLAS, 
TX 75235-8314
Employee Info
United States Department of The Treasury
Employee Id: 9999999998 IRS No. 000000000000
INTERNAL REVENUE SERVICE, 2021041800
PO BOX 1214, Rate Units Total YTD Taxes / Deductions Current YTD
CHARLOTTE, NC 28201-1214 - - $70,842,745,000.00 $70,842,745,000.00 Federal Withholding $0.00 $0.00
Earnings FICA - Social Security $0.00 $8,853.60
Commissions FICA - Medicare $0.00 $0.00
Employer Taxes
FUTA $0.00 $0.00
SUTA $0.00 $0.00
EIN: 61-1767
ID91:900037305581 SSN: 633441725
YTD Gross Gross
$70,842,745,000.00 $70,842,745,000.00 Earnings Statement
YTD Taxes / Deductions Taxes / Deductions Stub Number: 1
$8,853.60 $0.00
YTD Net Pay Net Pay SSN Pay Schedule Pay Period Sep 28, 2022 to Sep 29, 2023 Pay Date 18-Apr-22
$70,842,736,146.40 $70,842,745,000.00 XXX-XX-1725 Annually
CHECK DATE CHECK NUMBER
18-Apr-22 70842745,00000
THIS IS NOT A CHECK
CHECK AMOUNT
VOID
INTERNAL REVENUE SERVICE,PO BOX 1214,CHARLOTTE, NC 28201-1214
ALINE Pay, FSDD, ADPCheck, WGPS, Garnishment Services, EBTS, Benefit Services, Other Bank        Bank Address        Account Name        ABA        DDA        Collection Method JPMorgan Chase        One Chase Manhattan Plaza New York, NY 10005        ADP Tax Services        021000021        323269036        Reverse Wire Impound Deutsche Bank        60 Wall Street New York, NY 10005-2858        ADP Tax Services        021001033        00416217        Reverse Wire Impound Tax & 401(k) Bank        Bank Address        Account Name        ABA        DDA        Collection Method JPMorgan Chase        One Chase Manhattan Plaza New York, NY 10005        ADP Tax Services        021000021        9102628675        Reverse Wire Impound Deutsche Bank        60 Wall Street New York, NY 10005-2858        ADP Tax Services        021001033        00153170        Reverse Wire Impound Workers Compensation Bank        Bank Address        Account Name        ABA        DDA        Collection Method JPMorgan Chase        One Chase Manhattan Plaza New York, NY 10005        ADP Tax Services        021000021        304939315        Reverse Wire Impound NOTICE CLIENT acknowledges that if sufficient funds are not available by the date required pursuant to the foregoing provisions of this Agreement, (1) CLIENT will immediately become solely responsible for all tax deposits and filings, all employee wages, all wage garnishments, all CLIENT third- party payments (e.g., vendor payments) and all related penalties and interest due then and thereafter, (2) any and all ADP Services may, at ADP’s option, be immediately terminated, (3) neither BANK nor ADP will have any further obligation to CLIENT or any third party with respect to any such Services and (4) ADP may take such action as it deems appropriate to collect ADP’s Fees for Services. Client shall not initiate any ACH transactions utilizing ADP’s services that constitute International ACH transactions without first (i) notifying ADP of such IAT transactions in writing utilizing ADP’s Declaration of International ACH Transaction form (or such other form as directed by ADP) and (ii) complying with the requirements applicable to IAT transactions. ADP shall not be liable for any delay or failure in processing any ACH transaction due to Client’s failure to so notify ADP of Client’s IAT transactions or Client’s failure to comply with applicable IAT requirements. EXHIBIT 99.1 ZACHRY WOOD15 $76,033,000,000.00 20,642,000,000 18,936,000,000 18,525,000,000 17,930,000,000 15,227,000,000 11,247,000,000 6,959,000,000 6,836,000,000 10,671,000,000 7,068,000,000For Disclosure, Privacy Act, and Paperwork Reduction ActNotice, see separate instructions. $76,033,000,000.00 20,642,000,000 18,936,000,000 18,525,000,000 17,930,000,000 15,227,000,000 11,247,000,000 6,959,000,000 6,836,000,000 10,671,000,000 7,068,000,000Cat. No. 11320B $76,033,000,000.00 20,642,000,000 18,936,000,000 18,525,000,000 17,930,000,000 15,227,000,000 11,247,000,000 6,959,000,000 6,836,000,000 10,671,000,000 7,068,000,000Form 1040 (2021) $76,033,000,000.00 20,642,000,000 18,936,000,000Reported Normalized and Operating Income/ExpenseSupplemental SectionTotal Revenue as Reported, Supplemental $257,637,000,000.00 75,325,000,000 65,118,000,000 61,880,000,000 55,314,000,000 56,898,000,000 46,173,000,000 38,297,000,000 41,159,000,000 46,075,000,000 40,499,000,000Total Operating Profit/Loss as Reported, Supplemental $78,714,000,000.00 21,885,000,000 21,031,000,000 19,361,000,000 16,437,000,000 15,651,000,000 11,213,000,000 6,383,000,000 7,977,000,000 9,266,000,000 9,177,000,000Reported Effective Tax Rate $0.16 0.179 0.157 0.158 0.158 0.159 0.119 0.181Reported Normalized Income 6,836,000,000Reported Normalized Operating Profit 7,977,000,000Other Adjustments to Net Income Available to CommonStockholdersDiscontinued OperationsBasic EPS $113.88 31.15 28.44 27.69 26.63 22.54 16.55 10.21 9.96 15.49 10.2Basic EPS from Continuing Operations $113.88 31.12 28.44 27.69 26.63 22.46 16.55 10.21 9.96 15.47 10.2Basic EPS from Discontinued OperationsDiluted EPS $112.20 30.69 27.99 27.26 26.29 22.3 16.4 10.13 9.87 15.35 10.12Diluted EPS from Continuing Operations $112.20 30.67 27.99 27.26 26.29 22.23 16.4 10.13 9.87 15.33 10.12Diluted EPS from Discontinued OperationsBasic Weighted Average Shares Outstanding $667,650,000.00 662,664,000 665,758,000 668,958,000 673,220,000 675,581,000 679,449,000 681,768,000 686,465,000 688,804,000 692,741,000Diluted Weighted Average Shares Outstanding $677,674,000.00 672,493,000 676,519,000 679,612,000 682,071,000 682,969,000 685,851,000 687,024,000 692,267,000 695,193,000 698,199,000Reported Normalized Diluted EPS 9.87Basic EPS $113.88 31.15 28.44 27.69 26.63 22.54 16.55 10.21 9.96 15.49 10.2 1Diluted EPS $112.20 30.69 27.99 27.26 26.29 22.3 16.4 10.13 9.87 15.35 10.12Basic WASO $667,650,000.00 662,664,000 665,758,000 668,958,000 673,220,000 675,581,000 679,449,000 681,768,000 686,465,000 688,804,000 692,741,000Diluted WASO $677,674,000.00 672,493,000 676,519,000 679,612,000 682,071,000 682,969,000 685,851,000 687,024,000 692,267,000 695,193,000 698,199,000Fiscal year end September 28th., 2022. | USDFor Paperwork Reduction Act Notice, see the seperateInstructions.THIS NOTE IS LEGAL TENDERTENDERFOR ALL DEBTS, PUBLIC ANDPRIVATECurrent ValueUnappropriated, Affiliated, Securities, at Value.(1) For subscriptions, your payment method on file will be automatically charged monthly/annually at the then-current list price until you cancel. If you have a discount it will apply to the then-current list price until it expires. To cancel your subscription at any time, go to Account & Settings and cancel the subscription. (2) For one-time services, your payment method on file will reflect the charge in the amount referenced in this invoice. Terms, conditions, pricing, features, service, and support options are subject to change without notice.All dates and times are Pacific Standard Time (PST).INTERNAL REVENUE SERVICE, $20,210,418.00 PO BOX 1214, Rate Units Total YTD Taxes / Deductions Current YTD CHARLOTTE, NC 28201-1214 - - $70,842,745,000.00 $70,842,745,000.00 Federal Withholding $0.00 $0.00 Earnings FICA - Social Security $0.00 $8,853.60 Commissions FICA - Medicare $0.00 $0.00 Employer Taxes FUTA $0.00 $0.00 SUTA $0.00 $0.00 EIN: 61-1767ID91:900037305581 SSN: 633441725 YTD Gross Gross $70,842,745,000.00 $70,842,745,000.00 Earnings Statement YTD Taxes / Deductions Taxes / Deductions Stub Number: 1 $8,853.60 $0.00 YTD Net Pay net, pay. SSN Pay Schedule Paid Period Sep 28, 2022 to Sep 29, 2023 15-Apr-22 Pay Day 18-Apr-22 $70,842,736,146.40 $70,842,745,000.00 XXX-XX-1725 Annually Sep 28, 2022 to Sep 29, 2023 CHECK DATE CHECK NUMBER 001000 18-Apr-22 PAY TO THE : ZACHRY WOOD ORDER OF :Office of the 46th President Of The United States. 117th US Congress Seal Of The US Treasury Department, 1769 W.H.W. DC, US 2022. : INTERNAL REVENUE SERVICE,PO BOX 1214,CHARLOTTE, NC 28201-1214 CHECK AMOUNT $70,842,745,000.00 Pay ZACHRY.WOOD************ :NON-NEGOTIABLE : VOID AFTER 14 DAYS INTERNAL REVENUE SERVICE :000,000.00 $18,936,000,000.00 $18,525,000,000.00 $17,930,000,000.00 $15,227,000,000.00 $11,247,000,000.00 $6,959,000,000.00 $6,836,000,000.00 $10,671,000,000.00 $7,068,000,000.00 $76,033,000,000.00 $20,642,000,000.00 $18,936,000,000.00 $18,525,000,000.00 $17,930,000,000.00 $15,227,000,000.00 $11,247,000,000.00 $6,959,000,000.00 $6,836,000,000.00 $10,671,000,000.00 $7,068,000,000.00 $76,033,000,000.00 $20,642,000,000.00 $18,936,000,000.00 $18,525,000,000.00 $17,930,000,000.00 $15,227,000,000.00 $11,247,000,000.00 $6,959,000,000.00 $6,836,000,000.00 $10,671,000,000.00 $7,068,000,000.00 $76,033,000,000.00 $20,642,000,000.00 $18,936,000,000.00 $257,637,000,000.00 $75,325,000,000.00 $65,118,000,000.00 $61,880,000,000.00 $55,314,000,000.00 $56,898,000,000.00 $46,173,000,000.00 $38,297,000,000.00 $41,159,000,000.00 $46,075,000,000.00 $40,499,000,000.00 $78,714,000,000.00 $21,885,000,000.00 $21,031,000,000.00 $19,361,000,000.00 $16,437,000,000.00 $15,651,000,000.00 $11,213,000,000.00 $6,383,000,000.00 $7,977,000,000.00 $9,266,000,000.00 $9,177,000,000.00 $0.16 $0.18 $0.16 $0.16 $0.16 $0.16 $0.12 $0.18 $6,836,000,000.00 $7,977,000,000.00 $113.88 $31.15 $28.44 $27.69 $26.63 $22.54 $16.55 $10.21 $9.96 $15.49 $10.20 $113.88 $31.12 $28.44 $27.69 $26.63 $22.46 $16.55 $10.21 $9.96 $15.47 $10.20 $112.20 $30.69 $27.99 $27.26 $26.29 $22.30 $16.40 $10.13 $9.87 $15.35 $10.12 $112.20 $30.67 $27.99 $27.26 $26.29 $22.23 $16.40 $10.13 $9.87 $15.33 $10.12 $667,650,000.00 $662,664,000.00 $665,758,000.00 $668,958,000.00 $673,220,000.00 $675,581,000.00 $679,449,000.00 $681,768,000.00 $686,465,000.00 $688,804,000.00 $692,741,000.00 $677,674,000.00 $672,493,000.00 $676,519,000.00 $679,612,000.00 $682,071,000.00 $682,969,000.00 $685,851,000.00 $687,024,000.00 $692,267,000.00 $695,193,000.00 $698,199,000.00 $9.87 $113.88 $31.15 $28.44 $27.69 $26.63 $22.54 $16.55 $10.21 $9.96 $15.49 $10.20 $1.00 $112.20 $30.69 $27.99 $27.26 $26.29 $22.30 $16.40 $10.13 $9.87 $15.35 $10.12 $667,650,000.00 $662,664,000.00 $665,758,000.00 $668,958,000.00 $673,220,000.00 $675,581,000.00 $679,449,000.00 $681,768,000.00 $686,465,000.00 $688,804,000.00 $692,741,000.00 $677,674,000.00 $672,493,000.00 $676,519,000.00 $679,612,000.00 $682,071,000.00 $682,969,000.00 $685,851,000.00 $687,024,000.00 $692,267,000.00 $695,193,000.00 $698,199,000.00 : $70,842,745,000.00 633-44-1725 Annually : branches: - main : on: schedule: - cron: "0 2 * * 1-5 : obs: my_job: name :deploy to staging : runs-on :ubuntu-18.04 :The available virtual machine types are:ubuntu-latest, ubuntu-18.04, or ubuntu-16.04 :windows-latest :# :Controls when the workflow will run :"#":, "Triggers the workflow on push or pull request events but only for the "Masterbranch" branch :":, push: EFT information Routing number: 021000021Payment account ending: 9036Name on the account: ADPTax reporting informationInternal Revenue ServiceUnited States Department of the TreasuryMemphis, TN 375001-1498Tracking ID: 1023934415439Customer File Number: 132624428Date of Issue: 07-29-2022ZACHRY T WOOD3050 REMOND DR APT 1206DALLAS, TX 75211Taxpayer's Name: ZACH T WOOTaxpayer Identification Number: XXX-XX-1725Tax Period: December, 2018Return: 1040 ZACHRY TYLER WOOD 5323 BRADFORD DRIVE DALLAS TX 75235 EMPLOYER IDENTIFICATION NUMBER :611767919 :FIN :xxxxx4775 THE 101YOUR BASIC/DILUTED EPS RATE HAS BEEN CHANGED FROM $0.001 TO 33611.5895286 :State Income TaxTotal Work HrsBonusTrainingYour federal taxable wages this period are $22,756,988,716,000.00Net.Important Notes0.001 TO 112.20 PAR SHARE VALUETot*$70,842,743,866.00$22,756,988,716,000.00$22,756,988,716,000.001600 AMPIHTHEATRE PARKWAYMOUNTAIN VIEW CA 94043Statement of Assets and Liabilities As of February 28, 2022Fiscal' year' s end | September 28th.Total (includes tax of (00.00))   
+3/6/2022 at 6:37 PM
+Q4 2021 Q3 2021 Q2 2021 Q1 2021 Q4 2020
+GOOGL_income￾statement_Quarterly_As_Originally_Reported 24,934,000,000 25,539,000,000 37,497,000,000 31,211,000,000 30,818,000,000
+24,934,000,000 25,539,000,000 21,890,000,000 19,289,000,000 22,677,000,000
+Cash Flow from Operating Activities, Indirect 24,934,000,000 25,539,000,000 21,890,000,000 19,289,000,000 22,677,000,000
+Net Cash Flow from Continuing Operating Activities, Indirect 20,642,000,000 18,936,000,000 18,525,000,000 17,930,000,000 15,227,000,000
+Cash Generated from Operating Activities 6,517,000,000 3,797,000,000 4,236,000,000 2,592,000,000 5,748,000,000
+Income/Loss before Non-Cash Adjustment 3,439,000,000 3,304,000,000 2,945,000,000 2,753,000,000 3,725,000,000
+Total Adjustments for Non-Cash Items 3,439,000,000 3,304,000,000 2,945,000,000 2,753,000,000 3,725,000,000
+Depreciation, Amortization and Depletion, Non-Cash
+Adjustment 3,215,000,000 3,085,000,000 2,730,000,000 2,525,000,000 3,539,000,000
+Depreciation and Amortization, Non-Cash Adjustment 224,000,000 219,000,000 215,000,000 228,000,000 186,000,000
+Depreciation, Non-Cash Adjustment 3,954,000,000 3,874,000,000 3,803,000,000 3,745,000,000 3,223,000,000
+Amortization, Non-Cash Adjustment 1,616,000,000 -1,287,000,000 379,000,000 1,100,000,000 1,670,000,000
+Stock-Based Compensation, Non-Cash Adjustment -2,478,000,000 -2,158,000,000 -2,883,000,000 -4,751,000,000 -3,262,000,000
+Taxes, Non-Cash Adjustment -2,478,000,000 -2,158,000,000 -2,883,000,000 -4,751,000,000 -3,262,000,000
+Investment Income/Loss, Non-Cash Adjustment -14,000,000 64,000,000 -8,000,000 -255,000,000 392,000,000
+Gain/Loss on Financial Instruments, Non-Cash Adjustment -2,225,000,000 2,806,000,000 -871,000,000 -1,233,000,000 1,702,000,000
+Other Non-Cash Items -5,819,000,000 -2,409,000,000 -3,661,000,000 2,794,000,000 -5,445,000,000
+Changes in Operating Capital -5,819,000,000 -2,409,000,000 -3,661,000,000 2,794,000,000 -5,445,000,000
+Change in Trade and Other Receivables -399,000,000 -1,255,000,000 -199,000,000 7,000,000 -738,000,000
+Change in Trade/Accounts Receivable 6,994,000,000 3,157,000,000 4,074,000,000 -4,956,000,000 6,938,000,000
+Change in Other Current Assets 1,157,000,000 238,000,000 -130,000,000 -982,000,000 963,000,000
+Change in Payables and Accrued Expenses 1,157,000,000 238,000,000 -130,000,000 -982,000,000 963,000,000
+Change in Trade and Other Payables 5,837,000,000 2,919,000,000 4,204,000,000 -3,974,000,000 5,975,000,000
+Change in Trade/Accounts Payable 368,000,000 272,000,000 -3,000,000 137,000,000 207,000,000
+Change in Accrued Expenses -3,369,000,000 3,041,000,000 -1,082,000,000 785,000,000 740,000,000
+Change in Deferred Assets/Liabilities
+Change in Other Operating Capital
+-11,016,000,000 -10,050,000,000 -9,074,000,000 -5,383,000,000 -7,281,000,000
+Change in Prepayments and Deposits -11,016,000,000 -10,050,000,000 -9,074,000,000 -5,383,000,000 -7,281,000,000
+Cash Flow from Investing Activities
+Cash Flow from Continuing Investing Activities -6,383,000,000 -6,819,000,000 -5,496,000,000 -5,942,000,000 -5,479,000,000
+-6,383,000,000 -6,819,000,000 -5,496,000,000 -5,942,000,000 -5,479,000,000
+Purchase/Sale and Disposal of Property, Plant and Equipment,
+Net
+Purchase of Property, Plant and Equipment -385,000,000 -259,000,000 -308,000,000 -1,666,000,000 -370,000,000
+Sale and Disposal of Property, Plant and Equipment -385,000,000 -259,000,000 -308,000,000 -1,666,000,000 -370,000,000
+Purchase/Sale of Business, Net -4,348,000,000 -3,360,000,000 -3,293,000,000 2,195,000,000 -1,375,000,000
+Purchase/Acquisition of Business -40,860,000,000 -35,153,000,000 -24,949,000,000 -37,072,000,000 -36,955,000,000
+Purchase/Sale of Investments, Net
+Purchase of Investments 36,512,000,000 31,793,000,000 21,656,000,000 39,267,000,000 35,580,000,000
+100,000,000 388,000,000 23,000,000 30,000,000 -57,000,000
+Sale of Investments
+Other Investing Cash Flow -15,254,000,000
+Purchase/Sale of Other Non-Current Assets, Net -16,511,000,000 -15,254,000,000 -15,991,000,000 -13,606,000,000 -9,270,000,000
+Sales of Other Non-Current Assets -16,511,000,000 -12,610,000,000 -15,991,000,000 -13,606,000,000 -9,270,000,000
+Cash Flow from Financing Activities -13,473,000,000 -12,610,000,000 -12,796,000,000 -11,395,000,000 -7,904,000,000
+Cash Flow from Continuing Financing Activities 13,473,000,000 -12,796,000,000 -11,395,000,000 -7,904,000,000
+Issuance of/Payments for Common Stock, Net -42,000,000
+Payments for Common Stock 115,000,000 -42,000,000 -1,042,000,000 -37,000,000 -57,000,000
+Proceeds from Issuance of Common Stock 115,000,000 6,350,000,000 -1,042,000,000 -37,000,000 -57,000,000
+Issuance of/Repayments for Debt, Net 6,250,000,000 -6,392,000,000 6,699,000,000 900,000,000 0
+Issuance of/Repayments for Long Term Debt, Net 6,365,000,000 -2,602,000,000 -7,741,000,000 -937,000,000 -57,000,000
+Proceeds from Issuance of Long Term Debt
+Repayments for Long Term Debt 2,923,000,000 -2,453,000,000 -2,184,000,000 -1,647,000,000
+Proceeds from Issuance/Exercising of Stock Options/Warrants 0 300,000,000 10,000,000 3.38E+11
+Other Financing Cash Flow
+Cash and Cash Equivalents, End of Period
+Change in Cash 20,945,000,000 23,719,000,000 23,630,000,000 26,622,000,000 26,465,000,000
+Effect of Exchange Rate Changes 25930000000) 235000000000) -3,175,000,000 300,000,000 6,126,000,000
+Cash and Cash Equivalents, Beginning of Period PAGE="$USD(181000000000)".XLS BRIN="$USD(146000000000)".XLS 183,000,000 -143,000,000 210,000,000
+Cash Flow Supplemental Section $23,719,000,000,000.00 $26,622,000,000,000.00 $26,465,000,000,000.00 $20,129,000,000,000.00
+Change in Cash as Reported, Supplemental 2,774,000,000 89,000,000 -2,992,000,000 6,336,000,000
+Income Tax Paid, Supplemental 13,412,000,000 157,000,000
+ZACHRY T WOOD -4990000000
+Cash and Cash Equivalents, Beginning of Period
+Department of the Treasury
+Internal Revenue Service
+Q4 2020 Q4 2019
+Calendar Year
+Due: 04/18/2022
+Dec. 31, 2020 Dec. 31, 2019
+USD in "000'"s
+Repayments for Long Term Debt 182527 161857
+Costs and expenses:
+Cost of revenues 84732 71896
+Research and development 27573 26018
+Sales and marketing 17946 18464
+General and administrative 11052 9551
+European Commission fines 0 1697
+Total costs and expenses 141303 127626
+Income from operations 41224 34231
+Other income (expense), net 6858000000 5394
+Income before income taxes 22,677,000,000 19,289,000,000
+Provision for income taxes 22,677,000,000 19,289,000,000
+Net income 22,677,000,000 19,289,000,000
+*include interest paid, capital obligation, and underweighting
+Basic net income per share of Class A and B common stock
+and Class C capital stock (in dollars par share)
+Diluted net income per share of Class A and Class B common
+stock and Class C capital stock (in dollars par share)
+*include interest paid, capital obligation, and underweighting
+Basic net income per share of Class A and B common stock
+and Class C capital stock (in dollars par share)
+Diluted net income per share of Class A and Class B common
+stock and Class C capital stock (in dollars par share)
+ALPHABET 88-1303491
+5323 BRADFORD DR,
+DALLAS, TX 75235-8314
+Employee Info
+United States Department of The Treasury
+Employee Id: 9999999998 IRS No. 000000000000
+INTERNAL REVENUE SERVICE, 2021041800 :
