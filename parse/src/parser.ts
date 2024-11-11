import axios from "axios"
import { error, log } from "console";
import { readFileSync, writeFileSync } from "fs";
import pdf from 'pdf-parse';
import { writeHeapSnapshot } from "v8";
import { PdfReader } from "pdfreader";
import { execSync, spawnSync } from "child_process";

import http from 'https'; // or 'https' for https:// URLs
import fs from 'fs';
import { askOllama } from "./llm";                                                                                                              
import { cosineSimilarity, numberWithSpaces } from "./utils";

export async function parseLink(url: string, prefix: string) {
    // const req = await axios.get(url);

    const req = await axios.get("https://zakupki.mos.ru/newapi/api/Auction/Get?auctionId=" + url.slice(-7), {
        "headers": {
            
        },
    });   

    const filesRaw = req.data.files as {id:number, name:string}[];
    const filesLinks = filesRaw.map(file=>{
        return `https://zakupki.mos.ru/newapi/api/FileStorage/Download?id=${file.id}`
    })
    const files = filesRaw.map(file=>{
        return {
            ext: file.name.split('.').at(-1),
            mbTechTask: /(техническое[\-\_\s]?задание)|(т[3|з])/i.test(file.name),
            mbContractProject: /(проект[\-\_\s]?контракта)/i.test(file.name),
            ...file,
            textFilePath: ""
        }
    })

    // console.log(req.data)
    // console.log(filesLinks);
    // console.log(files);

    try{

        let i=0;
        for(const file of files){
            files[i].textFilePath = await downloadAndConvertFile(filesLinks[i], file.name, prefix + '_out'+i)
            i++;
        }

    }catch(e){
        console.error("Cant parse files, e: ", e)
        return;   
    }

    let criterias = new Array(6).fill(0).map(z=>false);

    const cp = files.find(f=>f.mbContractProject)
    const tt = files.find(f=>f.mbTechTask)
    // console.log(cp,tt);

    // await new Promise(r=>setTimeout(r, 100000));
    //  

    const cpt = cp ? readFileSync(cp!.textFilePath).toString() : "";
    const ttt = tt ? readFileSync(tt!.textFilePath).toString() : "";
    
    
    // crit 1
    console.log(url, 'crit1', crit1(req.data, cpt, ttt));

    // crit 2 
    console.log(url, 'crit2', crit2(req.data, cpt, ttt));

    // downloadAndConvertFile(filesLinks[1], files[1].name);
    return;
    
}

function crit1(apiAnswer : any, contractProjectText: string, techTaskText: string) {

    const cpt5 = contractProjectText.split('\n').slice(0, 5);
    const ttt5 = techTaskText.split('\n').slice(0, 5);

    const shouldBeSimToRaw = apiAnswer.name;
    const shouldBeSimTo = shouldBeSimToRaw.split(' ').map((word:string)=>{
        if(word.length <= 4) return word;
        return word.slice(0, -2) 
    }).join(' ')


    const globalSim = Math.max(...[...cpt5, ...ttt5].map(line=>{
        const prepared = line.split(' ').map(word=>{
            if(word.length <= 4) return word;
            return word.slice(0, -2) 
        }).join(' ')

        // console.log(line, prepared, shouldBeSimTo,cosineSimilarity(shouldBeSimTo, prepared));        
        const issim = cosineSimilarity(shouldBeSimTo, prepared);
        return isNaN(issim) ? 0 : issim;
    }))

    return globalSim
}

function crit2(apiAnswer : any, contractProjectText: string, techTaskText: string ){
    if (!apiAnswer.isContractGuaranteeRequired) return true
    
    const shouldBeAmount = apiAnswer.contractGuaranteeAmount;

    const amountCheck1 = shouldBeAmount.toString()
    const amountCheck2 = numberWithSpaces(shouldBeAmount);

    function checkFor(hay: string){
        const m0 = Math.max(hay.indexOf(amountCheck1), hay.indexOf(amountCheck2))
        if(m0 === -1) return false

        const m = Math.max(hay.indexOf("обеспечение исполнения"), hay.indexOf("обеспечения исполнения"))

        return m !== -1 && Math.abs(m-m0) < 120
    }
    
    return checkFor(contractProjectText) || checkFor(techTaskText)
}

async function downloadAndConvertFile(fileLink: string, fileName: string, outName: string) {
    console.log('req', fileLink, fileName);

    const lclCopy = process.env.CUSTOM_TEMP + outName + fileName.slice(fileName.lastIndexOf('.')); 

    await downloadFile(fileLink, lclCopy)    

    if(fileName.endsWith(".pdf")){  
        execSync(process.env.PDFTOTEXT_PATH! + " " + lclCopy + " "+outName+".txt", {cwd: process.env.CUSTOM_TEMP})
    }else if(fileName.endsWith(".docx") || fileName.endsWith(".doc")){
        const buf = execSync(process.env.LIBREOFFICE_PATH! + ` --convert-to "txt:Text (encoded):UTF8" "${lclCopy}" --outdir ${process.env.CUSTOM_TEMP}`)        
    }else{
        throw new Error("Unsupported file extension")
    }

    fs.rmSync(lclCopy)

    return process.env.CUSTOM_TEMP + outName + ".txt"
}

async function downloadFile(fileLink: string, toFile: string) {
    return new Promise((r,j)=>{

        const file = fs.createWriteStream(toFile);
        const request = http.get(fileLink, function(response) {
            response.pipe(file);
    
            file.on("finish", () => {
                file.close();
                r(true)
                console.log("Download Completed");
            });
        });
    
    })
}

