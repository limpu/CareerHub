// Universal A4 Export Engine for PDF, DOCX, and TXT (AT-002, FND-010 Compliant)

export interface ResumeExportProfile {
  fullName: string;
  headline: string;
  email: string;
  phone: string;
  location: string;
  summary: string;
  skills: string[];
  experiences: Array<{
    title: string;
    company: string;
    location: string;
    period: string;
    highlights: string[];
    skills?: string[];
  }>;
  education: Array<{
    degree: string;
    institution: string;
    year: string;
  }>;
  certificates?: Array<{
    name: string;
    issuer: string;
    issue_date: string;
  }>;
}

export interface CoverLetterExportData {
  recipientName: string;
  company: string;
  salutation: string;
  openingParagraph: string;
  bodyParagraphs: string[];
  closingParagraph: string;
  signoff: string;
}

// Exact ISO A4 dimensions in PDF points (72 pt/inch)
const A4_WIDTH = 595.28;
const A4_HEIGHT = 841.89;
const MARGIN_LEFT = 45;
const MARGIN_RIGHT = 45;
const MARGIN_TOP = 50;
const MARGIN_BOTTOM = 50;

function escapePDFText(text: string): string {
  return text.replace(/\\/g, '\\\\').replace(/\(/g, '\\(').replace(/\)/g, '\\)');
}

function wrapText(text: string, maxCharsPerLine: number = 88): string[] {
  const words = text.split(/\s+/);
  const lines: string[] = [];
  let curLine = '';

  for (const w of words) {
    if (!curLine) {
      curLine = w;
    } else if (curLine.length + 1 + w.length <= maxCharsPerLine) {
      curLine += ' ' + w;
    } else {
      lines.push(curLine);
      curLine = w;
    }
  }
  if (curLine) lines.push(curLine);
  return lines.length > 0 ? lines : [''];
}

function approxWidth(text: string, fontSize: number, bold: boolean = false): number {
  const factor = bold ? 0.58 : 0.52;
  return text.length * fontSize * factor;
}

export function buildProfessionalA4PDF(profile: ResumeExportProfile): Uint8Array {
  const pages: Array<{ commands: string[] }> = [];
  let currentCommands: string[] = [];
  let curY = A4_HEIGHT - MARGIN_TOP;

  const startNewPage = () => {
    if (currentCommands.length > 0) {
      pages.push({ commands: currentCommands });
    }
    currentCommands = [];
    curY = A4_HEIGHT - MARGIN_TOP;
  };

  const ensureSpace = (requiredHeight: number) => {
    if (curY - requiredHeight < MARGIN_BOTTOM) {
      startNewPage();
    }
  };

  const addText = (
    text: string,
    x: number,
    y: number,
    font: 'F1' | 'F2' = 'F1',
    size: number = 10,
    color: [number, number, number] = [0.1, 0.1, 0.1]
  ) => {
    currentCommands.push(
      `q ${color[0].toFixed(2)} ${color[1].toFixed(2)} ${color[2].toFixed(2)} rg BT /${font} ${size} Tf ${x.toFixed(2)} ${y.toFixed(2)} Td (${escapePDFText(text)}) Tj ET Q`
    );
  };

  const addLine = (
    x1: number,
    y1: number,
    x2: number,
    y2: number,
    color: [number, number, number] = [0.8, 0.8, 0.8],
    width: number = 0.5
  ) => {
    currentCommands.push(
      `q ${color[0].toFixed(2)} ${color[1].toFixed(2)} ${color[2].toFixed(2)} RG ${width} w ${x1.toFixed(2)} ${y1.toFixed(2)} m ${x2.toFixed(2)} ${y2.toFixed(2)} l S Q`
    );
  };

  // 1. Header: Candidate Info
  if (profile.fullName) {
    ensureSpace(60);
    const nameStr = profile.fullName.toUpperCase();
    const nameWidth = approxWidth(nameStr, 17, true);
    const nameX = Math.max(MARGIN_LEFT, (A4_WIDTH - nameWidth) / 2);
    addText(nameStr, nameX, curY, 'F2', 17, [0.05, 0.05, 0.1]);
    curY -= 16;

    if (profile.headline) {
      const hlWidth = approxWidth(profile.headline, 10, true);
      const hlX = Math.max(MARGIN_LEFT, (A4_WIDTH - hlWidth) / 2);
      addText(profile.headline, hlX, curY, 'F2', 10, [0.15, 0.38, 0.92]); // Blue #2563EB
      curY -= 14;
    }

    const contactParts: string[] = [];
    if (profile.email) contactParts.push(profile.email);
    if (profile.phone) contactParts.push(profile.phone);
    if (profile.location) contactParts.push(profile.location);

    if (contactParts.length > 0) {
      const contactStr = contactParts.join('   |   ');
      const cWidth = approxWidth(contactStr, 9, false);
      const cX = Math.max(MARGIN_LEFT, (A4_WIDTH - cWidth) / 2);
      addText(contactStr, cX, curY, 'F1', 9, [0.35, 0.40, 0.45]);
      curY -= 14;
    }

    addLine(MARGIN_LEFT, curY, A4_WIDTH - MARGIN_RIGHT, curY, [0.85, 0.88, 0.90], 1);
    curY -= 18;
  }

  const renderSectionHeader = (title: string) => {
    ensureSpace(35);
    addText(title.toUpperCase(), MARGIN_LEFT, curY, 'F2', 10.5, [0.05, 0.05, 0.1]);
    curY -= 4;
    addLine(MARGIN_LEFT, curY, A4_WIDTH - MARGIN_RIGHT, curY, [0.80, 0.82, 0.85], 0.75);
    curY -= 14;
  };

  // 2. Professional Summary
  if (profile.summary) {
    renderSectionHeader('Professional Summary');
    const summaryLines = wrapText(profile.summary, 90);
    for (const sLine of summaryLines) {
      ensureSpace(14);
      addText(sLine, MARGIN_LEFT, curY, 'F1', 9.5, [0.2, 0.2, 0.2]);
      curY -= 13;
    }
    curY -= 10;
  }

  // 3. Work Experience
  if (profile.experiences && profile.experiences.length > 0) {
    renderSectionHeader('Work Experience');
    for (const exp of profile.experiences) {
      ensureSpace(40);
      addText(exp.title, MARGIN_LEFT, curY, 'F2', 10, [0.05, 0.05, 0.1]);
      if (exp.period) {
        const pWidth = approxWidth(exp.period, 9, false);
        addText(exp.period, A4_WIDTH - MARGIN_RIGHT - pWidth, curY, 'F1', 9, [0.4, 0.45, 0.5]);
      }
      curY -= 13;

      const compLoc = exp.location ? `${exp.company}   |   ${exp.location}` : exp.company;
      addText(compLoc, MARGIN_LEFT, curY, 'F1', 9.5, [0.3, 0.35, 0.4]);
      curY -= 13;

      if (exp.highlights && exp.highlights.length > 0) {
        for (const h of exp.highlights) {
          const hLines = wrapText(h, 84);
          for (let idx = 0; idx < hLines.length; idx++) {
            ensureSpace(13);
            if (idx === 0) {
              addText('-', MARGIN_LEFT + 4, curY, 'F2', 9, [0.2, 0.2, 0.2]);
              addText(hLines[idx], MARGIN_LEFT + 14, curY, 'F1', 9, [0.25, 0.25, 0.25]);
            } else {
              addText(hLines[idx], MARGIN_LEFT + 14, curY, 'F1', 9, [0.25, 0.25, 0.25]);
            }
            curY -= 12;
          }
        }
      }
      curY -= 8;
    }
    curY -= 6;
  }

  // 4. Key Skills & Competencies
  if (profile.skills && profile.skills.length > 0) {
    renderSectionHeader('Key Skills & Competencies');
    const skillText = profile.skills.join('   |   ');
    const skillLines = wrapText(skillText, 88);
    for (const sl of skillLines) {
      ensureSpace(14);
      addText(sl, MARGIN_LEFT, curY, 'F1', 9.5, [0.2, 0.2, 0.2]);
      curY -= 13;
    }
    curY -= 10;
  }

  // 5. Education
  if (profile.education && profile.education.length > 0) {
    renderSectionHeader('Education');
    for (const edu of profile.education) {
      ensureSpace(30);
      addText(edu.degree, MARGIN_LEFT, curY, 'F2', 10, [0.05, 0.05, 0.1]);
      if (edu.year) {
        const yWidth = approxWidth(edu.year, 9, false);
        addText(edu.year, A4_WIDTH - MARGIN_RIGHT - yWidth, curY, 'F1', 9, [0.4, 0.45, 0.5]);
      }
      curY -= 13;
      addText(edu.institution, MARGIN_LEFT, curY, 'F1', 9.5, [0.3, 0.35, 0.4]);
      curY -= 16;
    }
    curY -= 4;
  }

  // 6. Certifications & Licenses
  if (profile.certificates && profile.certificates.length > 0) {
    renderSectionHeader('Certifications & Licenses');
    for (const cert of profile.certificates) {
      ensureSpace(20);
      const cTitle = `${cert.name} (${cert.issuer})`;
      addText(cTitle, MARGIN_LEFT, curY, 'F1', 9.5, [0.2, 0.2, 0.2]);
      if (cert.issue_date) {
        const dWidth = approxWidth(cert.issue_date, 9, false);
        addText(cert.issue_date, A4_WIDTH - MARGIN_RIGHT - dWidth, curY, 'F1', 9, [0.4, 0.45, 0.5]);
      }
      curY -= 14;
    }
  }

  if (currentCommands.length > 0) {
    pages.push({ commands: currentCommands });
  }

  const numPages = Math.max(1, pages.length);
  const encoder = new TextEncoder();

  const fontObj1 = 3 + numPages * 2;
  const fontObj2 = 4 + numPages * 2;

  const kidsArray: string[] = [];
  for (let i = 0; i < numPages; i++) {
    kidsArray.push(`${3 + i * 2} 0 R`);
  }

  let docStr = '%PDF-1.4\n%\xE2\xE3\xCF\xD3\n';
  const offsets: number[] = [];
  const getByteLen = (s: string) => encoder.encode(s).length;

  offsets.push(getByteLen(docStr));
  docStr += '1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n';

  offsets.push(getByteLen(docStr));
  docStr += `2 0 obj\n<< /Type /Pages /Kids [${kidsArray.join(' ')}] /Count ${numPages} >>\nendobj\n`;

  for (let i = 0; i < numPages; i++) {
    const pageObjNum = 3 + i * 2;
    const contentObjNum = 4 + i * 2;
    const streamContent = pages[i].commands.join('\n') + '\n';
    const streamLen = encoder.encode(streamContent).length;

    offsets.push(getByteLen(docStr));
    docStr += `${pageObjNum} 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 ${A4_WIDTH.toFixed(2)} ${A4_HEIGHT.toFixed(2)}] /Contents ${contentObjNum} 0 R /Resources << /Font << /F1 ${fontObj1} 0 R /F2 ${fontObj2} 0 R >> >> >>\nendobj\n`;

    offsets.push(getByteLen(docStr));
    docStr += `${contentObjNum} 0 obj\n<< /Length ${streamLen} >>\nstream\n${streamContent}endstream\nendobj\n`;
  }

  offsets.push(getByteLen(docStr));
  docStr += `${fontObj1} 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n`;

  offsets.push(getByteLen(docStr));
  docStr += `${fontObj2} 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold >>\nendobj\n`;

  const totalObjs = fontObj2 + 1;
  const xrefOffset = getByteLen(docStr);
  let xrefStr = `xref\n0 ${totalObjs}\n0000000000 65535 f \n`;
  for (const off of offsets) {
    xrefStr += off.toString().padStart(10, '0') + ' 00000 n \n';
  }
  docStr += xrefStr;
  docStr += `trailer\n<< /Size ${totalObjs} /Root 1 0 R >>\nstartxref\n${xrefOffset}\n%%EOF\n`;

  return encoder.encode(docStr);
}

// Cover Letter A4 PDF Generator
export function buildCoverLetterA4PDF(letter: CoverLetterExportData, candidate: { name: string; email: string; phone: string }): Uint8Array {
  const pages: Array<{ commands: string[] }> = [];
  let currentCommands: string[] = [];
  let curY = A4_HEIGHT - MARGIN_TOP;

  const ensureSpace = (h: number) => {
    if (curY - h < MARGIN_BOTTOM) {
      if (currentCommands.length > 0) pages.push({ commands: currentCommands });
      currentCommands = [];
      curY = A4_HEIGHT - MARGIN_TOP;
    }
  };

  const addText = (text: string, x: number, y: number, font: 'F1' | 'F2' = 'F1', size: number = 10, color: [number, number, number] = [0.1, 0.1, 0.1]) => {
    currentCommands.push(
      `q ${color[0].toFixed(2)} ${color[1].toFixed(2)} ${color[2].toFixed(2)} rg BT /${font} ${size} Tf ${x.toFixed(2)} ${y.toFixed(2)} Td (${escapePDFText(text)}) Tj ET Q`
    );
  };

  const addLine = (x1: number, y1: number, x2: number, y2: number) => {
    currentCommands.push(
      `q 0.85 0.88 0.90 RG 1 w ${x1.toFixed(2)} ${y1.toFixed(2)} m ${x2.toFixed(2)} ${y2.toFixed(2)} l S Q`
    );
  };

  // Header
  addText(candidate.name.toUpperCase(), MARGIN_LEFT, curY, 'F2', 15, [0.05, 0.05, 0.1]);
  curY -= 15;
  addText(`${candidate.email}   |   ${candidate.phone}`, MARGIN_LEFT, curY, 'F1', 9.5, [0.35, 0.40, 0.45]);
  curY -= 14;
  addLine(MARGIN_LEFT, curY, A4_WIDTH - MARGIN_RIGHT, curY);
  curY -= 20;

  // Date
  const dateStr = new Date().toLocaleDateString('en-US', { month: 'long', day: 'numeric', year: 'numeric' });
  addText(dateStr, MARGIN_LEFT, curY, 'F1', 9.5, [0.3, 0.3, 0.3]);
  curY -= 20;

  // Recipient
  addText(`To: ${letter.recipientName}`, MARGIN_LEFT, curY, 'F2', 10, [0.1, 0.1, 0.1]);
  curY -= 14;
  addText(letter.company, MARGIN_LEFT, curY, 'F1', 9.5, [0.3, 0.3, 0.3]);
  curY -= 24;

  // Salutation
  addText(letter.salutation, MARGIN_LEFT, curY, 'F2', 10, [0.1, 0.1, 0.1]);
  curY -= 18;

  // Opening
  const openLines = wrapText(letter.openingParagraph, 88);
  for (const line of openLines) {
    ensureSpace(14);
    addText(line, MARGIN_LEFT, curY, 'F1', 9.5, [0.15, 0.15, 0.15]);
    curY -= 14;
  }
  curY -= 10;

  // Body
  for (const bPara of letter.bodyParagraphs) {
    const bLines = wrapText(bPara, 88);
    for (const line of bLines) {
      ensureSpace(14);
      addText(line, MARGIN_LEFT, curY, 'F1', 9.5, [0.15, 0.15, 0.15]);
      curY -= 14;
    }
    curY -= 10;
  }

  // Closing
  const closeLines = wrapText(letter.closingParagraph, 88);
  for (const line of closeLines) {
    ensureSpace(14);
    addText(line, MARGIN_LEFT, curY, 'F1', 9.5, [0.15, 0.15, 0.15]);
    curY -= 14;
  }
  curY -= 16;

  // Signoff
  const signLines = letter.signoff.split('\n');
  for (const s of signLines) {
    ensureSpace(14);
    addText(s, MARGIN_LEFT, curY, 'F1', 9.5, [0.2, 0.2, 0.2]);
    curY -= 13;
  }

  if (currentCommands.length > 0) pages.push({ commands: currentCommands });

  const numPages = Math.max(1, pages.length);
  const encoder = new TextEncoder();
  const fontObj1 = 3 + numPages * 2;
  const fontObj2 = 4 + numPages * 2;
  const kidsArray: string[] = [];
  for (let i = 0; i < numPages; i++) kidsArray.push(`${3 + i * 2} 0 R`);

  let docStr = '%PDF-1.4\n%\xE2\xE3\xCF\xD3\n';
  const offsets: number[] = [];
  const getByteLen = (s: string) => encoder.encode(s).length;

  offsets.push(getByteLen(docStr));
  docStr += '1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n';
  offsets.push(getByteLen(docStr));
  docStr += `2 0 obj\n<< /Type /Pages /Kids [${kidsArray.join(' ')}] /Count ${numPages} >>\nendobj\n`;

  for (let i = 0; i < numPages; i++) {
    const pageObjNum = 3 + i * 2;
    const contentObjNum = 4 + i * 2;
    const streamContent = pages[i].commands.join('\n') + '\n';
    const streamLen = encoder.encode(streamContent).length;

    offsets.push(getByteLen(docStr));
    docStr += `${pageObjNum} 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 ${A4_WIDTH.toFixed(2)} ${A4_HEIGHT.toFixed(2)}] /Contents ${contentObjNum} 0 R /Resources << /Font << /F1 ${fontObj1} 0 R /F2 ${fontObj2} 0 R >> >> >>\nendobj\n`;
    offsets.push(getByteLen(docStr));
    docStr += `${contentObjNum} 0 obj\n<< /Length ${streamLen} >>\nstream\n${streamContent}endstream\nendobj\n`;
  }

  offsets.push(getByteLen(docStr));
  docStr += `${fontObj1} 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n`;
  offsets.push(getByteLen(docStr));
  docStr += `${fontObj2} 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold >>\nendobj\n`;

  const totalObjs = fontObj2 + 1;
  const xrefOffset = getByteLen(docStr);
  let xrefStr = `xref\n0 ${totalObjs}\n0000000000 65535 f \n`;
  for (const off of offsets) xrefStr += off.toString().padStart(10, '0') + ' 00000 n \n';
  docStr += xrefStr;
  docStr += `trailer\n<< /Size ${totalObjs} /Root 1 0 R >>\nstartxref\n${xrefOffset}\n%%EOF\n`;

  return encoder.encode(docStr);
}

// DOCX Helpers
function escapeXML(str: string): string {
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&apos;');
}

function makeCrcTable(): Uint32Array {
  const table = new Uint32Array(256);
  for (let n = 0; n < 256; n++) {
    let c = n;
    for (let k = 0; k < 8; k++) {
      c = c & 1 ? 0xedb88320 ^ (c >>> 1) : c >>> 1;
    }
    table[n] = c;
  }
  return table;
}

const CRC_TABLE = makeCrcTable();

function crc32(bytes: Uint8Array): number {
  let c = 0xffffffff;
  for (let i = 0; i < bytes.length; i++) {
    c = CRC_TABLE[(c ^ bytes[i]) & 0xff] ^ (c >>> 8);
  }
  return (c ^ 0xffffffff) >>> 0;
}

function createZipArchive(files: Array<{ name: string; content: string }>): Uint8Array {
  const encoder = new TextEncoder();
  const fileEntries: Array<{ nameBytes: Uint8Array; dataBytes: Uint8Array; crc: number; offset: number }> = [];
  const chunks: Uint8Array[] = [];
  let currentOffset = 0;

  for (const f of files) {
    const nameBytes = encoder.encode(f.name);
    const dataBytes = encoder.encode(f.content);
    const crc = crc32(dataBytes);

    const localHeader = new Uint8Array(30);
    const dv = new DataView(localHeader.buffer);
    dv.setUint32(0, 0x04034b50, true);
    dv.setUint16(4, 20, true);
    dv.setUint16(6, 0, true);
    dv.setUint16(8, 0, true);
    dv.setUint16(10, 0, true);
    dv.setUint16(12, 0, true);
    dv.setUint32(14, crc, true);
    dv.setUint32(18, dataBytes.length, true);
    dv.setUint32(22, dataBytes.length, true);
    dv.setUint16(26, nameBytes.length, true);
    dv.setUint16(28, 0, true);

    chunks.push(localHeader, nameBytes, dataBytes);
    fileEntries.push({ nameBytes, dataBytes, crc, offset: currentOffset });
    currentOffset += localHeader.length + nameBytes.length + dataBytes.length;
  }

  const centralDirStart = currentOffset;
  let centralDirSize = 0;

  for (const entry of fileEntries) {
    const cdHeader = new Uint8Array(46);
    const dv = new DataView(cdHeader.buffer);
    dv.setUint32(0, 0x02014b50, true);
    dv.setUint16(4, 20, true);
    dv.setUint16(6, 20, true);
    dv.setUint16(8, 0, true);
    dv.setUint16(10, 0, true);
    dv.setUint16(12, 0, true);
    dv.setUint16(14, 0, true);
    dv.setUint32(16, entry.crc, true);
    dv.setUint32(20, entry.dataBytes.length, true);
    dv.setUint32(24, entry.dataBytes.length, true);
    dv.setUint16(28, entry.nameBytes.length, true);
    dv.setUint16(30, 0, true);
    dv.setUint16(32, 0, true);
    dv.setUint16(34, 0, true);
    dv.setUint16(36, 0, true);
    dv.setUint32(38, 0, true);
    dv.setUint32(42, entry.offset, true);

    chunks.push(cdHeader, entry.nameBytes);
    centralDirSize += cdHeader.length + entry.nameBytes.length;
  }

  const eocd = new Uint8Array(22);
  const dvEocd = new DataView(eocd.buffer);
  dvEocd.setUint32(0, 0x06054b50, true);
  dvEocd.setUint16(4, 0, true);
  dvEocd.setUint16(6, 0, true);
  dvEocd.setUint16(8, fileEntries.length, true);
  dvEocd.setUint16(10, fileEntries.length, true);
  dvEocd.setUint32(12, centralDirSize, true);
  dvEocd.setUint32(16, centralDirStart, true);
  dvEocd.setUint16(20, 0, true);

  chunks.push(eocd);

  const totalLength = chunks.reduce((acc, c) => acc + c.length, 0);
  const result = new Uint8Array(totalLength);
  let pos = 0;
  for (const c of chunks) {
    result.set(c, pos);
    pos += c.length;
  }
  return result;
}

export function buildProfessionalA4Docx(profile: ResumeExportProfile): Blob {
  const pList: string[] = [];

  if (profile.fullName) {
    pList.push(`    <w:p>
      <w:pPr><w:jc w:val="center"/><w:spacing w:after="40"/></w:pPr>
      <w:r><w:rPr><w:b/><w:sz w:val="34"/><w:color w:val="0F172A"/></w:rPr><w:t>${escapeXML(profile.fullName.toUpperCase())}</w:t></w:r>
    </w:p>`);

    if (profile.headline) {
      pList.push(`    <w:p>
        <w:pPr><w:jc w:val="center"/><w:spacing w:after="40"/></w:pPr>
        <w:r><w:rPr><w:b/><w:sz w:val="22"/><w:color w:val="2563EB"/></w:rPr><w:t>${escapeXML(profile.headline)}</w:t></w:r>
      </w:p>`);
    }

    const contactParts: string[] = [];
    if (profile.email) contactParts.push(profile.email);
    if (profile.phone) contactParts.push(profile.phone);
    if (profile.location) contactParts.push(profile.location);

    if (contactParts.length > 0) {
      pList.push(`    <w:p>
        <w:pPr>
          <w:jc w:val="center"/>
          <w:spacing w:after="160"/>
          <w:pBdr><w:bottom w:val="single" w:sz="12" w:space="8" w:color="CBD5E1"/></w:pBdr>
        </w:pPr>
        <w:r><w:rPr><w:sz w:val="18"/><w:color w:val="64748B"/></w:rPr><w:t>${escapeXML(contactParts.join('   |   '))}</w:t></w:r>
      </w:p>`);
    }
  }

  const addHeading = (title: string) => {
    pList.push(`    <w:p>
      <w:pPr>
        <w:spacing w:before="240" w:after="80"/>
        <w:pBdr><w:bottom w:val="single" w:sz="6" w:space="4" w:color="94A3B8"/></w:pBdr>
      </w:pPr>
      <w:r><w:rPr><w:b/><w:sz w:val="22"/><w:color w:val="0F172A"/></w:rPr><w:t>${escapeXML(title.toUpperCase())}</w:t></w:r>
    </w:p>`);
  };

  if (profile.summary) {
    addHeading('Professional Summary');
    pList.push(`    <w:p>
      <w:pPr><w:spacing w:after="120"/><w:jc w:val="both"/></w:pPr>
      <w:r><w:rPr><w:sz w:val="20"/><w:color w:val="334155"/></w:rPr><w:t>${escapeXML(profile.summary)}</w:t></w:r>
    </w:p>`);
  }

  if (profile.experiences && profile.experiences.length > 0) {
    addHeading('Work Experience');
    for (const exp of profile.experiences) {
      pList.push(`    <w:p>
        <w:pPr>
          <w:spacing w:before="120" w:after="20"/>
          <w:tabs><w:tab w:val="right" w:pos="9600"/></w:tabs>
        </w:pPr>
        <w:r><w:rPr><w:b/><w:sz w:val="21"/><w:color w:val="0F172A"/></w:rPr><w:t>${escapeXML(exp.title)}</w:t></w:r>
        ${exp.period ? `<w:r><w:tab/><w:rPr><w:sz w:val="18"/><w:color w:val="64748B"/></w:rPr><w:t>${escapeXML(exp.period)}</w:t></w:r>` : ''}
      </w:p>`);

      const compLoc = exp.location ? `${exp.company}   |   ${exp.location}` : exp.company;
      pList.push(`    <w:p>
        <w:pPr><w:spacing w:after="60"/></w:pPr>
        <w:r><w:rPr><w:i/><w:sz w:val="19"/><w:color w:val="475569"/></w:rPr><w:t>${escapeXML(compLoc)}</w:t></w:r>
      </w:p>`);

      if (exp.highlights && exp.highlights.length > 0) {
        for (const h of exp.highlights) {
          pList.push(`    <w:p>
            <w:pPr>
              <w:ind w:left="280" w:hanging="200"/>
              <w:spacing w:after="40"/>
            </w:pPr>
            <w:r><w:rPr><w:b/><w:sz w:val="19"/><w:color w:val="334155"/></w:rPr><w:t>- </w:t></w:r>
            <w:r><w:rPr><w:sz w:val="19"/><w:color w:val="334155"/></w:rPr><w:t>${escapeXML(h)}</w:t></w:r>
          </w:p>`);
        }
      }
    }
  }

  if (profile.skills && profile.skills.length > 0) {
    addHeading('Key Skills & Competencies');
    pList.push(`    <w:p>
      <w:pPr><w:spacing w:after="120"/></w:pPr>
      <w:r><w:rPr><w:sz w:val="20"/><w:color w:val="334155"/></w:rPr><w:t>${escapeXML(profile.skills.join('   |   '))}</w:t></w:r>
    </w:p>`);
  }

  if (profile.education && profile.education.length > 0) {
    addHeading('Education');
    for (const edu of profile.education) {
      pList.push(`    <w:p>
        <w:pPr>
          <w:spacing w:before="80" w:after="20"/>
          <w:tabs><w:tab w:val="right" w:pos="9600"/></w:tabs>
        </w:pPr>
        <w:r><w:rPr><w:b/><w:sz w:val="21"/><w:color w:val="0F172A"/></w:rPr><w:t>${escapeXML(edu.degree)}</w:t></w:r>
        ${edu.year ? `<w:r><w:tab/><w:rPr><w:sz w:val="18"/><w:color w:val="64748B"/></w:rPr><w:t>${escapeXML(edu.year)}</w:t></w:r>` : ''}
      </w:p>`);
      pList.push(`    <w:p>
        <w:pPr><w:spacing w:after="80"/></w:pPr>
        <w:r><w:rPr><w:sz w:val="19"/><w:color w:val="475569"/></w:rPr><w:t>${escapeXML(edu.institution)}</w:t></w:r>
      </w:p>`);
    }
  }

  if (profile.certificates && profile.certificates.length > 0) {
    addHeading('Certifications & Licenses');
    for (const cert of profile.certificates) {
      pList.push(`    <w:p>
        <w:pPr>
          <w:spacing w:before="60" w:after="40"/>
          <w:tabs><w:tab w:val="right" w:pos="9600"/></w:tabs>
        </w:pPr>
        <w:r><w:rPr><w:sz w:val="19"/><w:color w:val="334155"/></w:rPr><w:t>${escapeXML(cert.name)} (${escapeXML(cert.issuer)})</w:t></w:r>
        ${cert.issue_date ? `<w:r><w:tab/><w:rPr><w:sz w:val="18"/><w:color w:val="64748B"/></w:rPr><w:t>${escapeXML(cert.issue_date)}</w:t></w:r>` : ''}
      </w:p>`);
    }
  }

  const docXml = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
${pList.join('\n')}
    <w:sectPr>
      <w:pgSz w:w="11906" w:h="16838"/>
      <w:pgMar w:top="1152" w:right="1152" w:bottom="1152" w:left="1152"/>
    </w:sectPr>
  </w:body>
</w:document>`;

  const contentTypesXml = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`;

  const relsXml = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`;

  const zipBytes = createZipArchive([
    { name: '[Content_Types].xml', content: contentTypesXml },
    { name: '_rels/.rels', content: relsXml },
    { name: 'word/document.xml', content: docXml },
  ]);

  return new Blob([zipBytes.buffer as ArrayBuffer], {
    type: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  });
}

// Cover Letter DOCX Generator
export function buildCoverLetterDocx(letter: CoverLetterExportData, candidate: { name: string; email: string; phone: string }): Blob {
  const pList: string[] = [];

  // Header
  pList.push(`    <w:p>
    <w:pPr><w:spacing w:after="40"/></w:pPr>
    <w:r><w:rPr><w:b/><w:sz w:val="30"/><w:color w:val="0F172A"/></w:rPr><w:t>${escapeXML(candidate.name.toUpperCase())}</w:t></w:r>
  </w:p>`);

  pList.push(`    <w:p>
    <w:pPr>
      <w:spacing w:after="160"/>
      <w:pBdr><w:bottom w:val="single" w:sz="12" w:space="8" w:color="CBD5E1"/></w:pBdr>
    </w:pPr>
    <w:r><w:rPr><w:sz w:val="18"/><w:color w:val="64748B"/></w:rPr><w:t>${escapeXML(`${candidate.email}   |   ${candidate.phone}`)}</w:t></w:r>
  </w:p>`);

  // Date
  pList.push(`    <w:p>
    <w:pPr><w:spacing w:after="140"/></w:pPr>
    <w:r><w:rPr><w:sz w:val="19"/><w:color w:val="475569"/></w:rPr><w:t>${escapeXML(new Date().toLocaleDateString('en-US', { month: 'long', day: 'numeric', year: 'numeric' }))}</w:t></w:r>
  </w:p>`);

  // Recipient
  pList.push(`    <w:p>
    <w:pPr><w:spacing w:after="40"/></w:pPr>
    <w:r><w:rPr><w:b/><w:sz w:val="20"/><w:color w:val="0F172A"/></w:rPr><w:t>To: ${escapeXML(letter.recipientName)}</w:t></w:r>
  </w:p>`);
  pList.push(`    <w:p>
    <w:pPr><w:spacing w:after="180"/></w:pPr>
    <w:r><w:rPr><w:sz w:val="19"/><w:color w:val="475569"/></w:rPr><w:t>${escapeXML(letter.company)}</w:t></w:r>
  </w:p>`);

  // Salutation
  pList.push(`    <w:p>
    <w:pPr><w:spacing w:after="120"/></w:pPr>
    <w:r><w:rPr><w:b/><w:sz w:val="20"/><w:color w:val="0F172A"/></w:rPr><w:t>${escapeXML(letter.salutation)}</w:t></w:r>
  </w:p>`);

  // Opening
  pList.push(`    <w:p>
    <w:pPr><w:spacing w:after="120"/><w:jc w:val="both"/></w:pPr>
    <w:r><w:rPr><w:sz w:val="20"/><w:color w:val="1E293B"/></w:rPr><w:t>${escapeXML(letter.openingParagraph)}</w:t></w:r>
  </w:p>`);

  // Body
  for (const bPara of letter.bodyParagraphs) {
    pList.push(`    <w:p>
      <w:pPr><w:spacing w:after="120"/><w:jc w:val="both"/></w:pPr>
      <w:r><w:rPr><w:sz w:val="20"/><w:color w:val="1E293B"/></w:rPr><w:t>${escapeXML(bPara)}</w:t></w:r>
    </w:p>`);
  }

  // Closing
  pList.push(`    <w:p>
    <w:pPr><w:spacing w:after="180"/><w:jc w:val="both"/></w:pPr>
    <w:r><w:rPr><w:sz w:val="20"/><w:color w:val="1E293B"/></w:rPr><w:t>${escapeXML(letter.closingParagraph)}</w:t></w:r>
  </w:p>`);

  // Signoff
  const signLines = letter.signoff.split('\n');
  for (const s of signLines) {
    pList.push(`    <w:p>
      <w:pPr><w:spacing w:after="30"/></w:pPr>
      <w:r><w:rPr><w:sz w:val="19"/><w:color w:val="334155"/></w:rPr><w:t>${escapeXML(s)}</w:t></w:r>
    </w:p>`);
  }

  const docXml = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
${pList.join('\n')}
    <w:sectPr>
      <w:pgSz w:w="11906" w:h="16838"/>
      <w:pgMar w:top="1152" w:right="1152" w:bottom="1152" w:left="1152"/>
    </w:sectPr>
  </w:body>
</w:document>`;

  const contentTypesXml = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`;

  const relsXml = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`;

  const zipBytes = createZipArchive([
    { name: '[Content_Types].xml', content: contentTypesXml },
    { name: '_rels/.rels', content: relsXml },
    { name: 'word/document.xml', content: docXml },
  ]);

  return new Blob([zipBytes.buffer as ArrayBuffer], {
    type: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  });
}

export function buildTxtBlob(lines: string[]): Blob {
  return new Blob([lines.join('\r\n')], { type: 'text/plain;charset=utf-8' });
}

export function downloadBlob(blob: Blob, fileName: string) {
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = fileName;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}
