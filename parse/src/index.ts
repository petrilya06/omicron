import Fastify from 'fastify'
import { parseLink } from './parser';
import 'dotenv/config'
import { askOllama } from './llm';
import { readFileSync } from 'fs';
import { cosineSimilarity } from './utils';
import process from 'node:process';

const fastify = Fastify({
  logger: true
})

fastify.post('/parseLinks/', async function handler (request, reply) {

    const data = request.body
    if (!data || typeof data !== 'string') {
        reply.code(418);
        reply.send("Nonono mister fish, give me links in text/plain format divided by comma")
        return;
    }

    for (const link of data.split(',')){
        await parseLink(link, link.slice(link.lastIndexOf('/')+1))
    }

    return { hello: data }
});

(async()=>{

    console.log(cosineSimilarity(`Постав военн техни мебе д кабинет`, "МЕБЕ УЧЕНИЧЕСК"))
    parseLink("https://zakupki.mos.ru/auction/9864533", "9864533_");
    // askOllama(readFileSync('D:\\Temp\\out0.txt').toString());
    try {
        await fastify.listen({ port: 52052 })
    } catch (err) {
        fastify.log.error(err)
        process.exit(1)
    }
})()
