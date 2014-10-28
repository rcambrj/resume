#!/usr/bin/env node

// input
const markdownFile = `${__dirname}/README.md`;
// output
const markupFile = `${__dirname}/README.html`;
const pdfFile = `${__dirname}/README.pdf`;

const fs = require('fs');
const axios = require('axios');
const process = require('process');
const getCssFromGithub = require('generate-github-markdown-css');
const puppeteer = require('puppeteer');

const readFile = async (filename) => {
  return new Promise((resolve, reject) => {
    fs.readFile(filename, (err, text) => {
      if (err) {
        reject(new Error(`Unable to read file ${filename}: ${err}`));
      }
      resolve(text);
    })
  })
}

const writeFile = async (filename, contents) => {
  return new Promise((resolve, reject) => {
    fs.writeFile(filename, contents, (err) => {
      if (err) {
        reject(new Error(`Unable to write file ${filename}: ${err}`));
      }
      resolve()
    })
  });
}

const getMarkupFromGithub = async (markdown) => {
  // https://docs.github.com/en/rest/reference/markdown#render-a-markdown-document
  try {
    const response = await axios({
        url: 'https://api.github.com/markdown',
        method: 'post',
        data: {
          mode: 'markdown',
          text: `${markdown}`
        },
    });

    return response.data;
  } catch (err) {
    throw new Error(`Failed to fetch from github: ${err.response.status} ${JSON.stringify(err.response.data)}`);
  }
}

const writePdfFromHtml = async (pdfFilename, html) => {
  // https://github.com/puppeteer/puppeteer/blob/main/docs/api.md#pagepdfoptions
  const browser = await puppeteer.launch();
  const page = await browser.newPage();
  await page.setContent(`${html}`, {
    waitUntil: 'networkidle2'
  });
  await page.pdf({
    path: pdfFilename,
    format: 'A4',
    margin: {
      top: '50px',
      bottom: '50px',
      right: '30px',
      left: '30px',
    }
  });

  await browser.close();
}

const run = async () => {
  try {
    const markdown = await readFile(markdownFile);
    const [ markup, css ] = await Promise.all([
      getMarkupFromGithub(markdown),
      getCssFromGithub(),
    ])

    const wrappedMarkup =
    `<!DOCTYPE html><html><head><meta charset="utf-8">`+
    `<style>
      @media print {
        .markdown-body {
          width: 980px;
        }
      }
      .markdown-body {
        box-sizing: border-box;
        min-width: 200px;
        max-width: 980px;
        margin: 0 auto;
        padding: 15px;
      }
      ${css}
    </style>`+
    `</head><body>`+
    `<div class="markdown-body">${markup}</div>`+
    `</body></html>`;

    await Promise.all([
      writeFile(markupFile, wrappedMarkup),
      writePdfFromHtml(pdfFile, wrappedMarkup),
    ]);

    process.exit(0);
  } catch (err) {
    console.error(`${err}`);
    process.exit(1);
  }
}

run();