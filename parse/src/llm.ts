import { readFileSync } from 'fs'
import ollama from 'ollama'

export async function askOllama(fileContent: string) {

    console.time("gen")
    console.log(123);
    
 
    const response = await ollama.chat({
        model: 'gemma2',
        messages: [
            { role: 'system', content: readFileSync('prompts/test.txt').toString()},
            { role: 'user', content: fileContent}
            // { role: 'system', content: ""},
            // { role: 'user', content: "Why is sky blue? In two sentences."}
        ],
        
        options: {
            'temperature': 0.9
        }
    })


    console.log(response.message.content);
    

    console.timeEnd("gen")

}